Get GeoIP City database
-----------------------

sudo apt install geoipupdate
sudo cp /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
sudo mkdir -p /usr/local/share/GeoIP
sudo geoipupdate -v
