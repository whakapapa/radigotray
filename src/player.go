package main

//TODO undone



#include <memory>
#include <thread>

#include <glib.h>

#include <gst/gstversion.h>

#include <gstreamermm.h>
#include <gstreamermm/playbin.h>

#include <glibmm.h>


#if GST_VERSION_MAJOR >= 1
typedef Gst::PlayBin PlayBin;
#else
typedef Gst::PlayBin2 PlayBin;
#endif

class Player
{
	public:
	Player() = default;
	Player(const Player&) = delete;

	bool init(int argc, char** argv);
	void play(const Glib::ustring& url, const Glib::ustring& station = Glib::ustring());
	void play();
	void pause();
	void stop();
	void start();

	Glib::ustring get_station();
	bool has_station();
	Glib::RefPtr<PlayBin> get_playbin();

	bool init_streams(const Glib::ustring& data_url, const Glib::ustring& station);
	void set_config(const std::shared_ptr<Config>& cfg);

	std::shared_ptr<EventManager> em;

	private:
	std::shared_ptr<Config> config;
	Glib::RefPtr<PlayBin> playbin;

	Playlist playlist;
	MediaStreams streams;
	MediaStreams::iterator next_stream;

	Glib::ustring current_station;
	bool buffering = false;

	bool on_bus_message(const Glib::RefPtr<Gst::Bus>& bus, const Glib::RefPtr<Gst::Message>& message);
	void set_stream(const Glib::ustring& url);
	void set_buffer();
	void play_next_stream();
	};




	bool
	Player::init(int argc, char** argv)
	{
		Gst::init(argc, argv);

		playbin = PlayBin::create();
		if (!playbin) {
			LOG(ERROR) << "The PlayBin element could not be created.";
			return false;
		}
		set_buffer();

		Glib::RefPtr<Gst::Bus> bus = playbin->get_bus();
		bus->add_watch(sigc::mem_fun(*this, &Player::on_bus_message));

		return playlist.init();
	}



	void
	Player::play(const Glib::ustring& url, const Glib::ustring& station)
	{
		auto ok = init_streams(url, station);
		if (ok) {
			play();
		}
	}

	void
	Player::play()
	{
		if (streams.empty()) {
			LOG(DEBUG) << "Streams are empty!";
			return;
		}

		Glib::ustring stream_url = streams.front();
		next_stream = std::next(std::begin(streams));

		stop();
		set_stream(stream_url);
		start();
	}

	void
	Player::pause()
	{
		playbin->set_state(Gst::STATE_PAUSED);
	}

	void
	Player::stop()
	{
		playbin->set_state(Gst::STATE_NULL);
	}

	void
	Player::start()
	{
		playbin->set_state(Gst::STATE_PLAYING);
	}

	Glib::RefPtr<PlayBin>
	Player::get_playbin()
	{
		return playbin;
	}

	void
	Player::set_stream(const Glib::ustring& url)
	{
		playbin->property_uri() = url;
	}

	void
	Player::set_buffer()
	{
		playbin->property_buffer_size() = config->buffer_size * config->bufferDuration;
		playbin->property_buffer_duration() = config->bufferDuration * GST_SECOND;
	}

	void
	Player::play_next_stream()
	{
		auto stream_found = false;

		stop();

		while (next_stream != std::end(streams) and (not stream_found)) {
			auto u = *next_stream;
			next_stream++;

			if (gst_uri_is_valid(u.c_str()) != 0) {
				LOG(DEBUG) << "Trying to play stream: " << u;

				set_buffer();
				set_stream(u);
				start();

				stream_found = true;
			}
		}
	}

	bool
	Player::on_bus_message(const Glib::RefPtr<Gst::Bus>& /*bus*/, const Glib::RefPtr<Gst::Message>& message)
	{
		auto message_type = message->get_message_type();

		if (message_type == Gst::MESSAGE_EOS) {
			play_next_stream();
			} else if (message_type == Gst::MESSAGE_ERROR) {
				auto error_msg = Glib::RefPtr<Gst::MessageError>::cast_static(message);
				Glib::ustring e = "Error";

				if (error_msg) {
					#if GSTREAMERMM_MAJOR_VERSION == 1 and GSTREAMERMM_MINOR_VERSION >= 8
					Glib::Error err = error_msg->parse_error();
					#else
					Glib::Error err = error_msg->parse();
					#endif
					e.append(": ").append(err.what());
				}

				LOG(ERROR) << e;
				em->broadcast_info_changed(current_station, e);

				play_next_stream();
				} else if (message_type == Gst::MESSAGE_TAG) {
					auto msg_tag = Glib::RefPtr<Gst::MessageTag>::cast_static(message);
					Gst::TagList tag_list;
					#if GST_VERSION_MAJOR >= 1
					#if GSTREAMERMM_MAJOR_VERSION == 1 and GSTREAMERMM_MINOR_VERSION >= 8
					tag_list = msg_tag->parse_tag_list();
					#else
					msg_tag->parse(tag_list);
					#endif
					#else
					Glib::RefPtr<Gst::Pad> pad;
					msg_tag->parse(pad, tag_list);
					#endif
					if (tag_list.exists("title") && tag_list.size("title") > 0) {
						Glib::ustring title;
						auto ok = tag_list.get("title", title);
						if (ok) {
							em->broadcast_info_changed(current_station, title);
						}
					}
					} else if (message_type == Gst::MESSAGE_STATE_CHANGED) {
						auto state_changed_msg = Glib::RefPtr<Gst::MessageStateChanged>::cast_static(message);

						if (playbin->get_name() == state_changed_msg->get_source()->get_name()) {

							#if GSTREAMERMM_MAJOR_VERSION == 1 and GSTREAMERMM_MINOR_VERSION >= 8
							Gst::State new_state = state_changed_msg->parse_new_state();
							Gst::State old_state = state_changed_msg->parse_old_state();
							#else
							Gst::State new_state = state_changed_msg->parse();
							Gst::State old_state = state_changed_msg->parse_old();
							#endif

							tStationState st;
							if (new_state == Gst::State::STATE_PLAYING) {
								st = tStationState::ssPlaying
								} else {
									st = tStationState::ssIdle
								}

								em->state_changed(current_station, st);
								em->state = st;

								auto print = [](Gst::State& state) -> std::string {
									switch (state) {
									case Gst::State::STATE_PLAYING:
										return "STATE_PLAYING";
									case Gst::State::STATE_NULL:
										return "STATE_NULL";
									case Gst::State::STATE_READY:
										return "STATE_READY";
									case Gst::State::STATE_PAUSED:
										return "STATE_PAUSED";
									case Gst::State::STATE_VOID_PENDING:
										return "STATE_VOID_PENDING";
									default:
										return "STATE_UNKNOWN";
									}
									};

									LOG(DEBUG) << "Type: Gst::MESSAGE_STATE_CHANGED."
									<< " Old: " << print(old_state) << " New: " << print(new_state)
									<< " Source: " << state_changed_msg->get_source()->get_name();
								}
								} else if (message_type == Gst::MESSAGE_BUFFERING) {
									auto buffering_msg = Glib::RefPtr<Gst::MessageBuffering>::cast_static(message);
									#if GSTREAMERMM_MAJOR_VERSION == 1 and GSTREAMERMM_MINOR_VERSION >= 8
									auto percent = buffering_msg->parse_buffering();
									#else
									auto percent = buffering_msg->parse();
									#endif

									if (percent == 100) {
										buffering = false;
										playbin->set_state(Gst::STATE_PLAYING);
										LOG(DEBUG) << "buffering done";
										} else {
											buffering = true;

											Gst::State state, pending;

											playbin->get_state(state, pending, Gst::CLOCK_TIME_NONE);
											if (state != Gst::STATE_PAUSED) {
												playbin->set_state(Gst::STATE_PAUSED);
											}

											std::stringstream ss;

											ss << "buffering " << percent << "%";
											em->broadcast_info_changed(current_station, ss.str());

											LOG(DEBUG) << "buffering: " << percent;
										}
									}

									return true;
								}

								Glib::ustring
								Player::get_station()
								{
									return current_station;
								}

								bool
								Player::has_station()
								{
									return (!current_station.empty());
								}

								bool
								Player::init_streams(const Glib::ustring& data_url, const Glib::ustring& station)
								{
									bool ok;
									MediaStreams new_streams;

									std::tie(ok, new_streams) = playlist.get_streams(data_url);
									if ((not ok) or new_streams.empty()) {
										em->state_changed(current_station, tStationState::ssIdle)
										em->state = tStationState::ssIdle
										em->broadcast_info_changed(current_station, "Error: couldn't get audio stream!");
										LOG(ERROR) << "Couldn't get audio streams!";
										return false;
										} else {
											current_station = station;
											streams = new_streams;
										}

										return true;
									}

									void
									Player::set_config(const std::shared_ptr<Config>& cfg)
									{
										config = cfg;
										playlist.set_config(cfg);
									}
