.PHONY: install uninstall

install:
	wget -P /etc/systemd/system/ https://raw.githubusercontent.com/high-moctane/milbot/master/milbot.service
	systemctl enable milbot.service

uninstall:
	systemctl disable milbot.service
	rm -f /etc/systemd/system/milbot.service
