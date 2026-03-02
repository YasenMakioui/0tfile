# 0tfile
NOTE: Currently building, this uploads the file and returns a hash than can be used for download. But it doesn't encrypt the file and does not generate any secret

Send ephimeral and encrypted files to recipients

A super simple go app to send files using onetime urls.

Upload a file to /f and get a download url and a onetime secret to decrypt it.

It doesn't use a database, instead it creates metadata files to keep track of them.

A max download count and date limit can be used when uploading files.

Example usage:

```shell
curl -X POST -F "file=@./main.go" -H "max-download-count:3" -H "max-upload-days:5" http://localhost:3000/f
```

This will return 4 things:

* Download URL
* Secret URL
* Deletion token for the onetime secret
* Deletion token for the file

The posibility to upload without encryption is also posible adding the encrypt header to false.

Headers:

* max-download-count - int
* max-upload-days - int
* encrypt - bool 

After arriving to the max download or time limit, the file will appear as deleted.

A cron for file removal can be used executing the cleanup.sh script to clean expired files.
