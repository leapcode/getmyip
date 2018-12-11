Get GeoIP City database
-----------------------

You can use ``geoipupdate`` to download MaxMind's City database::

    sudo apt install geoipupdate
    sudo cp /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
    sudo geoipupdate -v

Usage
-----------------------

-geodb <path>
	path to the GeoLite2-City database (default is "/var/lib/GeoIP/GeoLite2-City.mmdb")
-port <port>
	port where the service listens on (default is 9001)

