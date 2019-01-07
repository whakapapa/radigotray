package decoder


#include <algorithm>
#include <cctype>
#include <functional>
#include <iostream>
#include <locale>
#include <memory>
#include <string>
#include <vector>

#include "easyloggingpp/easylogging++.h"


using MediaStreams = std::vector<std::string>;

type tDecoder uint8
const (
	DecUnknown = tDecoder(iota)
	DecM3U
	DecPLS
	DecRAM
	DecASX
	DecXSPF
)

type iDecoder interface {
	isValid() bool
	desciptMedia() string
	extractMedia() MediaStreams
	trimMedia()
}

func decValid(decoder iDecoder) bool {

}

func decDesc(decoder iDecoder) string {

}

func decExtract(decoder iDecoder) MediaStreams {

}

func decTrim(decoder iDecoder) {

}


//TODO OLD code here
class PlaylistDecoder
{


	virtual bool is_valid(const std::string& content_type) const = 0;
	virtual MediaStreams extract_media_streams(const std::string& data) = 0;
	virtual std::string desc() const = 0;

	protected:
	void trim(std::string& s)
	{
		s.erase(s.begin(), std::find_if(s.begin(), s.end(), std::not1(std::ptr_fun<int, int>(std::isspace))));
		s.erase(std::find_if(s.rbegin(), s.rend(), std::not1(std::ptr_fun<int, int>(std::isspace))).base(), s.end());
	}
}
