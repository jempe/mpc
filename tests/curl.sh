if [ -z "$BASEURL" ]; then
	BASEURL="http://localhost:4000"
fi

echo "BASEURL: $BASEURL"

if [ -z "$USERNAME" ]; then
	USERNAME="admin"
fi

if [ -z "$PASSWORD" ]; then
	PASSWORD="password"
fi

echo "USERNAME: $USERNAME"
echo "PASSWORD: $PASSWORD"

curl --request GET \
  --url $BASEURL/v1/videos \
  -u $USERNAME:$PASSWORD | jq
