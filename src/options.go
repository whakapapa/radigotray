package main

//TODO undone



import "./decoder"



#include <iomanip>
#include <sstream>
#include <iostream>
#include <memory>
#include <getopt.h>



class tDecXSPF : public PlaylistDecoder
{
	bool is_valid(const std::string& content_type) const override;
	MediaStreams extract_media_streams(const std::string& data) override;
	std::string desc() const override;


	};



	class CmdLineOptions
	{
		public:
		CmdLineOptions() = default;

		bool parse(int argc, char** argv);
		void show_help();

		bool resume = false;
		bool help = false;
		};




		// clang-format off
		static const struct option longopts[] = {
			{ "resume", no_argument, nullptr, 'r' },
			{ "help", no_argument, nullptr, 'h' },
			{ nullptr, 0, nullptr, 0 }
			};
			// clang-format on

			bool
			CmdLineOptions::parse(int argc, char** argv)
			{
				int opt, optidx;

				while ((opt = getopt_long(argc, argv, "rh", longopts, &optidx)) != -1) {
					switch (opt) {
					case 'r':
						resume = true;
						break;
					case 'h':
						help = true;
						break;
					default:
						return false;
					}
				}

				return true;
			}

			void
			CmdLineOptions::show_help()
			{
				std::cout << "Online radio streaming player" << std::endl;
				std::cout << "Usage:" << std::endl;
				std::cout << " radiotray-go [OPTIONS...]" << std::endl << std::endl;
				std::cout << " -h, --help show this help and exit" << std::endl;
				std::cout << " -r, --resume resume last played station on startup" << std::endl;
				std::cout << std::endl;
			}
