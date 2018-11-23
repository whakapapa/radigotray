package main


import "./base"
import "./decoder"


#include <algorithm>
#include <map>
#include <string>
#include <tuple>
#include <vector>

#include <curl/curl.h>
#include <magic.h>


class Playlist
{
	public:
	Playlist();
	Playlist(const Playlist&) = delete;

	~Playlist();

	bool init();
	std::tuple<bool, MediaStreams> get_streams(const std::string& url);
	void set_config(const std::shared_ptr<Config>& cfg);

	private:
	CURL* handle = nullptr;
	char errbuffer[CURL_ERROR_SIZE];

	magic_t mcookie = nullptr;

	std::shared_ptr<Config> config;

	bool abort_get_request = false;
	std::string data;

	std::map<tDecoder, std::shared_ptr<PlaylistDecoder>> decoders;

	void prepare_playlist_request(const std::string& url, bool only_headers);
	long get_http_status();
	std::string get_content_type();
	MediaStreams run_playlist_decoders(const std::string& url);
	bool has_prefix(const std::string& prefix, const std::string& str);
	tDecoder guess_playlist_decoder_type();

	static size_t write_memory_cb(void* ptr, size_t size, size_t nmemb, void* data);
	};


	static const char* cUserAgent = "Radio Tray Go (+https://github.com/whakapapa/radiotray-go)"
	static const bool cOnlyHeaders = true

	Playlist::Playlist()
	{
		decoders.emplace(std::make_pair(tDecoder::DecM3U, std::make_shared<M3UPlaylistDecoder>()));
		decoders.emplace(std::make_pair(tDecoder::DecPLS, std::make_shared<PLSPlaylistDecoder>()));
		decoders.emplace(std::make_pair(tDecoder::DecASX, std::make_shared<ASXPlaylistDecoder>()));
		decoders.emplace(std::make_pair(tDecoder::DecRAM, std::make_shared<tDecRAM>()));
		decoders.emplace(std::make_pair(tDecoder::DecXSPF, std::make_shared<tDecXSPF>()));
	}

	Playlist::~Playlist()
	{
		curl_easy_cleanup(handle);

		if (mcookie != nullptr) {
			magic_close(mcookie);
		}
	}

	bool
	Playlist::init()
	{
		handle = curl_easy_init();
		curl_easy_setopt(handle, CURLOPT_ERRORBUFFER, errbuffer);

		mcookie = magic_open(MAGIC_NONE);
		if (mcookie == nullptr) {
			LOG(ERROR) << "Error opening libmagic database";
			return false;
		}

		auto rc = magic_load(mcookie, nullptr);
		if (rc != 0) {
			LOG(ERROR) << magic_error(mcookie);
			return false;
		}

		return true;
	}

	std::tuple<bool, MediaStreams>
	Playlist::get_streams(const std::string& url)
	{
		MediaStreams streams;

		if (has_prefix("mms://", url)) {
			streams.push_back(url);
			return std::make_tuple(true, streams);
		}

		bool status = false;

		prepare_playlist_request(url, cOnlyHeaders);
		auto rc = curl_easy_perform(handle);
		auto httpStatusCode = get_http_status();

		if (rc == CURLE_OK and httpStatusCode == cHTTPok) {
			streams = run_playlist_decoders(url);
			status = true;
			} else {
				// Some streaming services don't allow HEAD request so we have
				// to do GET request and receive some small amount of data.
				LOG(WARNING) << "HEAD request failed!";

				abort_get_request = true;
				prepare_playlist_request(url, not cOnlyHeaders);
				rc = curl_easy_perform(handle);
				abort_get_request = false;

				if (rc == CURLE_OK or rc == CURLE_WRITE_ERROR /* it's ok, we've aborted reading */) {
					streams = run_playlist_decoders(url);
					status = true;
				}
			}

			return std::make_tuple(status, streams);
		}

		void
		Playlist::set_config(const std::shared_ptr<Config>& cfg)
		{
			config = cfg;
		}

		void
		Playlist::prepare_playlist_request(const std::string& url, bool only_headers)
		{
			data.clear();
			curl_easy_reset(handle);

			curl_easy_setopt(handle, CURLOPT_NOSIGNAL, 1);
			curl_easy_setopt(handle, CURLOPT_TIMEOUT_MS, config->urlTimeoutMS);

			curl_easy_setopt(handle, CURLOPT_FOLLOWLOCATION, 1);
			curl_easy_setopt(handle, CURLOPT_MAXREDIRS, 7);

			curl_easy_setopt(handle, CURLOPT_WRITEFUNCTION, write_memory_cb);
			curl_easy_setopt(handle, CURLOPT_WRITEDATA, static_cast<void*>(this));

			curl_easy_setopt(handle, CURLOPT_USERAGENT, cUserAgent);

			curl_easy_setopt(handle, CURLOPT_ENCODING, "gzip, deflate");

			curl_easy_setopt(handle, CURLOPT_URL, url.c_str());

			if (only_headers) {
				curl_easy_setopt(handle, CURLOPT_NOBODY, 1);
			}
		}

		long
		Playlist::get_http_status()
		{
			long status;

			auto rc = curl_easy_getinfo(handle, CURLINFO_RESPONSE_CODE, &status);
			if (rc != 0) {
				status = 0;
			}

			return status;
		}

		std::string
		Playlist::get_content_type()
		{
			char* content_type_data;
			std::string content_type;

			auto rc = curl_easy_getinfo(handle, CURLINFO_CONTENT_TYPE, &content_type_data);
			if (rc == CURLE_OK && content_type_data != nullptr) {
				content_type = content_type_data;
				for (auto& v : content_type) {
					v = std::tolower(v);
				}
			}

			return content_type;
		}

		MediaStreams
		Playlist::run_playlist_decoders(const std::string& url)
		{
			MediaStreams streams;
			bool extracted = false;

			// First try to detect playlist's type by Content-Type header
			auto content_type = get_content_type();

			if (not content_type.empty()) {
				LOG(DEBUG) << "Content-Type: " << content_type;
				for (const auto& decoder : decoders) {
					if (decoder.second->is_valid(content_type)) {
						LOG(DEBUG) << "Matched " << decoder.second->desc();
						prepare_playlist_request(url, not cOnlyHeaders);
						auto rc = curl_easy_perform(handle);
						if (rc == CURLE_OK) {
							LOG(DEBUG) << "Playlist: " << data;
							streams = decoder.second->extract_media_streams(data);
							for (auto& s : streams) {
								LOG(DEBUG) << "Stream: " << s;
							}
							extracted = true;
							break;
						}
					}
				}
			}

			// Try to infer playlist's type
			abort_get_request = true;
			prepare_playlist_request(url, not cOnlyHeaders);
			auto rc = curl_easy_perform(handle);
			abort_get_request = false;

			if (rc == CURLE_OK or rc == CURLE_WRITE_ERROR /* it's ok, we've aborted reading */) {
				auto type = guess_playlist_decoder_type();
				if (type != tDecoder::DecUnknown) {
					const auto& decoder = decoders[type];
					streams = decoder->extract_media_streams(data);
					extracted = true;
				}
			}

			// No decoder found, consider url a media stream
			if (streams.empty() and not extracted) {
				streams.push_back(url);
			}

			return streams;
		}

		bool
		Playlist::has_prefix(const std::string& prefix, const std::string& str)
		{
			if (prefix.size() > str.size()) {
				return false;
			}

			return (strncasecmp(str.c_str(), prefix.c_str(), prefix.size()) == 0);
		}

		tDecoder
		Playlist::guess_playlist_decoder_type()
		{
			auto desc = magic_buffer(mcookie, data.c_str(), data.size());
			if (desc == nullptr) {
				return tDecoder::DecUnknown
			}

			LOG(DEBUG) << "libmagic description: " << desc;

			static const std::string kM3UPlaylistDesc = "M3U";
			static const std::string kPLSPlaylistDesc = "PLS";
			static const std::string kXSPFPlaylistDesc = "XML";

			if (strncasecmp(desc, kM3UPlaylistDesc.c_str(), kM3UPlaylistDesc.size()) == 0) {
				return tDecoder::DecM3U
			}

			if (strncasecmp(desc, kPLSPlaylistDesc.c_str(), kPLSPlaylistDesc.size()) == 0) {
				return tDecoder::DecPLS
			}

			if (strncasecmp(desc, kXSPFPlaylistDesc.c_str(), kXSPFPlaylistDesc.size()) == 0) {
				std::transform(data.begin(), data.end(), data.begin(), tolower);
				auto pos = data.find("/xspf.org/");
				if (pos != std::string::npos) {
					return tDecoder::DecXSPF
				}
			}

			static const std::string kASXPlaylist = "<asx ";
			if (strncasecmp(data.c_str(), kASXPlaylist.c_str(), kASXPlaylist.size()) == 0) {
				return tDecoder::DecASX
			}

			return tDecoder::DecUnknown
		}

		size_t
		Playlist::write_memory_cb(void* ptr, size_t size, size_t nmemb, void* data)
		{
			size_t realsize = 0;
			auto instance = static_cast<Playlist*>(data);

			if (instance != nullptr) {
				realsize = size * nmemb;
				instance->data.append(static_cast<char*>(ptr), realsize);
				if (instance->abort_get_request) {
					realsize = 0;
				}
			}

			return realsize;
		}
