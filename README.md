# radio-ingest daemon

<p align="center">
  <img src="misc/logo.svg" width="200px" />
</p>


Server daemon handling incoming uploads on the radio UploadLink using the Omnia's [Notification Gateway](https://api.docs.nexx.cloud/notification-gateway). The application performs the following steps:

- Receiving calls form Omnia's Notification Gateway.
- Tries to map to information on show name and date given by the uploader to their respective fields.
- Sends a notification to the Stackfield channel.