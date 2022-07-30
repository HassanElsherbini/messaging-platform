# Messaging Platform
Initial work for a messaging platform prototype that supports sending/recieving messages and providing messaging analytics.


## Supported features
### Messaging
#### Facebook messenger
- Sending a [user feedback template](https://developers.facebook.com/docs/messenger-platform/send-messages/templates/customer-feedback-template) message to a user. (learn more about fb app modes and their levels of restriction when messaging users [here](https://developers.facebook.com/docs/development/build-and-test))
  ```bash
    POST /api/messaging/fbmessenger/send
    body: {
      "recipientId": "",
      "templateType": "" // currently only "customer_feedback" is supported
    }
  ```
- Receiving and persisting responses to feedback message templates. 
- Keeping track of which messages were read.

<br>

### Analytics web console
- Bar chart displaying count of messages sent, read and responses receieved for each day of the week.
- Bar chart displaying sentiment levels of responses. (number of satisfied, neutral and unsatisfied)


<br>

## Run Locally

Clone the project

```bash
  git clone https://github.com/HassanElsherbini/messaging-platform.git
```

API

1. Update .env.example file found in api/ with your own keys and rename it to .env.

2. A live public URL is needed to register our endpoint as a webhook with fb messenger, use [ngrok](https://ngrok.com/) to establish a secure tunnel to a live public endpoint.
```bash
  ngrok http 3000 //or your chosen server port
```

3. Start the server
```bash
  go run main.go
```

<br>

Client

1. Install dependencies
```bash
  npm install
```

2. Start dev server
```bash
  npm start
```
