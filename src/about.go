package main


import "./base"
#include <gtkmm.h>



class AboutDialog : public Gtk::AboutDialog
{
	public:
	AboutDialog();
	};

	//TODO work in progress



	AboutDialog::AboutDialog()
	{
		auto icon = std::string(cImagePath) + std::string(cAppIcon)

		this->set_icon_from_file(icon);
		this->set_program_name(cAppName)
		this->set_version(cAppVersion)

		auto logo = Gdk::Pixbuf::create_from_file(icon);
		this->set_logo(logo);

		std::vector<Glib::ustring> authors = { cAuthor }
		this->set_authors(authors);

		this->set_license_type(Gtk::License::LICENSE_GPL_3_0);

		this->set_website(cWeb)
		this->set_website_label("Project's website");

		char copyright[1024];
		memset(copyright, 0, sizeof(copyright));
		snprintf(copyright, sizeof(copyright) - 1, cCopyrightTmpl, cAppName, cCopyrightYear, cAuthor)
		this->set_copyright(copyright)
	}
