package decoder

//TODO undone

#include <iomanip>
#include <sstream>
#include "pugixml/pugixml.hpp"


type tDecXSPF struct {
	//TODO any variables for this struct?

}

func (dec *tDecXSPF) Decoder() {
	//TODO insert tDecXSPF methods here...

}


class tDecXSPF : public PlaylistDecoder
{
	bool is_valid(const std::string& content_type) const override;
	MediaStreams extract_media_streams(const std::string& data) override;
	std::string desc() const override;
	};



	func (decoder tDecXSPF) isValid(const std::string& content_type) bool {
		result := false

		if (content_type.find("application/xspf+xml") != std::string::npos) {
			result = true
		}

		return result
	}

	MediaStreams
	tDecXSPF::extract_media_streams(const std::string& data)
	{
		MediaStreams streams;

		pugi::xml_parse_result parsed = playlist_doc.load_buffer(data.c_str(), data.size());
		if (parsed) {
			try {
				pugi::xpath_node_set nodes = playlist_doc.select_nodes("//track/location");
				for (auto& node : nodes) {
					if (not node.node().text().empty()) {
						streams.emplace_back(node.node().text().as_string());
					}
				}
				} catch (pugi::xpath_exception& exc) {
					LOG(ERROR) << "Parsing XSPF playlist failed: " << exc.what();
				}
				} else {
					LOG(ERROR) << "Parsing XSPF playlist failed: " << parsed.description();
				}

				return streams;
			}

			std::string
			tDecXSPF::desc() const
			{
				return std::string("XSPF playlist decoder");
			}
