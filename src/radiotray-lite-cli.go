package main


#include <chrono>
#include <iostream>
#include <thread>

#include <glibmm.h>
#include <gstreamermm.h>

#include <stdlib.h>
#include <termios.h>



#include <memory>
#include <string>
#include <vector>

#include <stdlib.h>
#include <termios.h>

#include <gstreamermm.h>


class Player;

class KeyboardControl
{
	public:
	KeyboardControl(std::shared_ptr<Player> player, std::vector<std::string> stations)
	: player(player)
	, stations(stations)
	{
	}

	KeyboardControl() = delete;
	KeyboardControl(const KeyboardControl&) = delete;

	void operator()()
	{
		std::cout << "Press <space> to stop/resume playing." << std::endl;
		std::cout << "Press n/p to play next/previous station." << std::endl;

		size_t index = 0;
		auto stations_count = stations.size();

		player->play(stations[index]);

		while (true) {
			auto c = getch();

			if (c == kPauseCommand) {
				if (paused) {
					player->start();
					} else {
						player->stop();
					}
					paused = not paused;
					} else if (c == kNextStationCommand) {
						auto new_index = (stations_count > 1 ? (index + 1) % stations_count : index);
						if (new_index != index) {
							player->play(stations[new_index]);
							index = new_index;
						}
						} else if (c == kPreviousStationCommand) {
							if (stations_count > 1) {
								auto new_index = (index == 0 ? stations_count - 1 : (index - 1) % stations_count);
								if (new_index != index) {
									player->play(stations[new_index]);
									index = new_index;
								}
							}
						}
					}
				}

				private:
				std::shared_ptr<Player> player;
				std::vector<std::string> stations;

				static const int kPauseCommand = ' ';
				static const int kNextStationCommand = 'n';
				static const int kPreviousStationCommand = 'p';

				bool paused = false;

				int getch()
				{
					static struct termios oldt, newt;
					tcgetattr(STDIN_FILENO, &oldt); // save old settings
					newt = oldt;
					newt.c_lflag &= ~(ICANON); // disable buffering
					newt.c_lflag &= ~(ECHO); // disable echo
					tcsetattr(STDIN_FILENO, TCSANOW, &newt); // apply new settings

					int c = getchar(); // read character
					tcsetattr(STDIN_FILENO, TCSANOW, &oldt); // restore old settings

					return c;
				}
				};


				INITIALIZE_EASYLOGGINGPP

				void
				on_broadcast_info_changed_signal(Glib::ustring /*station*/, Glib::ustring info)
				{
					std::cout << "playing: " << info << std::endl;
				}

				int
				main(int argc, char** argv)
				{
					if (argc < 2) {
						std::cout << "Usage: " << argv[0] << " <uri>" << std::endl;
						return EXIT_FAILURE;
					}

					el::Configurations defaultConf;
					defaultConf.setToDefault();
					// Values are always std::string
					defaultConf.set(el::Level::Info, el::ConfigurationType::Format, "%datetime %level %loc %msg");
					defaultConf.set(el::Level::Error, el::ConfigurationType::Format, "%datetime %level %loc %msg");
					// default logger uses default configurations
					el::Loggers::reconfigureLogger("default", defaultConf);

					std::vector<std::string> stations;
					for (int i = 1; i < argc; i++) {
						stations.push_back(argv[i]);
					}

					auto config = std::make_shared<Config>();

					auto player = std::make_shared<Player>();
					player->em = std::make_shared<EventManager>(); // FIXME: EventManager should be part of Player?
					player->em->broadcast_info_changed.connect(sigc::ptr_fun(on_broadcast_info_changed_signal));
					player->set_config(config);

					auto ok = player->init(argc, argv);

					if (ok) {
						KeyboardControl keyboard(player, stations);
						std::thread t(std::ref(keyboard));

						auto mainloop = Glib::MainLoop::create();
						mainloop->run();
					}

					return EXIT_SUCCESS;
				}
