package decoder


#include <iomanip>
#include <sstream>


type tDecRAM struct {
	//TODO add properties here

}




{
	class tDecRAM : public PlaylistDecoder
	{


		bool is_valid(const std::string& content_type) const override;
		MediaStreams extract_media_streams(const std::string& data) override;
		std::string desc() const override;

		private:
		};


		bool
		tDecRAM::is_valid(const std::string& content_type) const
		{
			bool result = false;

			if (content_type.find("audio/x-pn-realaudio") != std::string::npos
			or content_type.find("audio/vnd.rn-realaudio") != std::string::npos) {
				result = true;
			}

			return result;
		}

		MediaStreams
		tDecRAM::extract_media_streams(const std::string& data)
		{
			MediaStreams streams;

			std::istringstream iss(data);
			std::string line;

			while (std::getline(iss, line)) {
				trim(line);
				if ((not line.empty()) and line.front() != '#') {
					streams.push_back(line);
				}
			}

			return streams;
		}

		std::string
		tDecRAM::desc() const
		{
			return std::string("RAM playlist decoder");
		}
