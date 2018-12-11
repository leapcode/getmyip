Get GeoIP City database
-----------------------

sudo apt install geoipupdate
sudo cp /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
sudo mkdir -p /usr/local/share/GeoIP
sudo geoipupdate -v

Usage
-----------------------

-geodb <path>
	path to the GeoLite2-City database (default is "/var/lib/GeoIP/GeoLite2-City.mmdb")
-port <port>
	port where the service listens on (default is 9001)

