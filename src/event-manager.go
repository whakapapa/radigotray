package main


#include <memory>
#include <string>

#include <gtkmm.h>
#include <sigc++/sigc++.h>


type tStationState uint8
const (
	ssUnknown = tStationState(iota)
	ssIdle
	ssConnecting
	ssPlaying
)

using BroadcastInfoChangedSignal = sigc::signal<void, Glib::ustring /*station*/, Glib::ustring /*info*/>;
using StateChangedSignal = sigc::signal<void, Glib::ustring /*station*/, tStationState /*state*/>;

class EventManager
{
	public:
	EventManager() = default;

	tStationState state = tStationState::ssUnknown

	BroadcastInfoChangedSignal broadcast_info_changed;
	StateChangedSignal state_changed;
	};

	std::string get_station_state_desc(tStationState state);





	std::string
	get_station_state_desc(tStationState state)
	{
		switch (state) {
		case tStationState::ssConnecting:
			return "ssConnecting"
		case tStationState::ssIdle:
			return "ssIdle"
		case tStationState::ssPlaying:
			return "ssPlaying"
		case tStationState::ssUnknown:
			return "ssUnknown"
		default:
			return "OOPS";
		}
	}
