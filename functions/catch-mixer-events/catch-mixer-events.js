const fetch = require('node-fetch');
const Mixer = require('@mixer/client-node');
const client = new Mixer.Client(new Mixer.DefaultRequestRunner());

// TODO for Chris: Add webhook URL for Discord channel
const WEBHOOK_URL = process.env.DISCORD_WEBHOOK_URL;

exports.handler = async (event, context) => {
  try {
    const jsonBody = JSON.parse(event.body);

    // We want to ignore all updates that don't represent online status
    if (jsonBody.payload.online) {
      console.log('********* streamer online!!!');
      const channelId = jsonBody.event.split(':')[1];

      client.use(
        new Mixer.OAuthProvider(client, {
          // TODO for Chris: Add Mixer Client ID
          clientId: ''
        })
      );

      const channelInfo = await client.request('GET', `channels/${channelId}`).then(res => {
        return res.body;
      });

      fetch(`${WEBHOOK_URL}`, {
        method: 'post',
        body: JSON.stringify({
          content: `${channelInfo.token} is now live! Check out the stream at https://www.mixer.com/${channelInfo.token}`
        }),
        headers: { 'Content-Type': 'application/json' }
      })
        .then(res => res.json())
        .then(json => console.log(json))
        .catch(err => console.error(err));
    }

    return {
      statusCode: 200,
      body: jsonBody.payload.online ? 'SUCCESS: Alert sent to Discord!' : 'No request sent...'
    };
  } catch (err) {
    return { statusCode: 500, body: `ERROR: ${err.toString()}` };
  }
};
