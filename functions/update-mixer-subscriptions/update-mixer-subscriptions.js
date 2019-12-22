// TODO Chris: Remove this imp;ort
const fetch = require('node-fetch');

const Mixer = require('@mixer/client-node');
const client = new Mixer.Client(new Mixer.DefaultRequestRunner());

client.use(
  new Mixer.OAuthProvider(client, {
    // TODO Chris: Swap these local variables with env variables
    clientId: AppSecrets.PARTY_CORGI_CLIENT_ID,
    secret: AppSecrets.PARTY_CORGI_SECRET
  })
);

exports.handler = async (event, context) => {
  try {
    // 1. Fetch list from /fetch-mixer-names
    const streamerList = await fetch(
      // TODO Chris: Swap this local variable with an env variable
      `${AppSecrets.SITE_URL}/.netlify/functions/fetch-mixer-names`,
      {
        method: 'post'
      }
    )
      .then(res => res.json())
      .then(data => data)
      .catch(err => console.log(err));

    const mixerUrls = streamerList.mixer;

    // 2. Get current hooks from Mixer
    const currentSubscriptions = await client
      .request('GET', `/hooks`, {
        headers: {
          // TODO Chris: Swap this local variable with an env variable
          Authorization: `Secret ${AppSecrets.PARTY_CORGI_SECRET}`
        }
      })
      .then(res => {
        return res.body.filter(sub => sub.isActive == true);
      });
    console.log('***** CURRENT SUBSCRIPTIONS');
    console.log(currentSubscriptions);

    // 3. Parse channel IDs from subscription data, create array
    const currentSubscriptionIds = currentSubscriptions.map(sub => {
      const broadcastEvent = sub.events.find(event => event.includes(':broadcast'));
      const channelIdRegex = new RegExp(/:(.*):/g);
      const idArr = channelIdRegex.exec(broadcastEvent);
      return idArr[1];
    });
    console.log('***** CURRENT SUBSCRIPTION IDS');
    console.log(currentSubscriptionIds);

    // 4. Get Mixer IDs from MD list based on URL
    const updatedSubscriptionIds = mixerUrls.map(async url => {
      // Parse username from URL
      const usernameRegex = new RegExp(/(\w+)$/gim);
      const usernameArr = Array.from(url.match(usernameRegex));

      // Get info about each user based on their user name and parse data
      return await client.request('GET', `/users/search?query=${usernameArr[0]}`).then(res => {
        const user = res.body[0];
        return {
          userId: user.id,
          username: user.username,
          channelId: user.channel.id
        };
      });
    });

    // Resolves all of the Promises created by the .map() call above
    const newChannelIds = await Promise.all(updatedSubscriptionIds)
      .then(subscriptions => {
        // 5. Check subscriptions, identify URLs to add
        return subscriptions.map(sub => {
          return !currentSubscriptionIds.includes(sub.channelId) ? sub.channelId.toString() : null;
        });
      })
      .catch(err => {
        console.log('Could not parse newChannelIds', err);
      });

    console.log('***** NEW CHANNEL IDS');
    console.log(newChannelIds);

    // 6. Add new subscriptions
    const idsToAdd = newChannelIds.filter(id => {
      return !currentSubscriptionIds.includes(id);
    });
    console.log('idsToAdd', idsToAdd);

    const addedIdsString = idsToAdd.join(', ');

    idsToAdd.map(id => {
      client
        .request(`POST`, `/hooks`, {
          body: {
            events: [`channel:${id}:broadcast`],
            kind: 'web',
            // TODO Chris: This URL should hit the deployed endpoint for the
            // catch-mixer-events function
            url: ``
          },
          headers: {
            // TODO Chris: Swap these local variables with env variables
            Authorization: `Secret ${AppSecrets.PARTY_CORGI_SECRET}`,
            'Client-ID': AppSecrets.PARTY_CORGI_CLIENT_ID
          }
        })
        .then(res => {
          console.log(res.body);
        })
        .catch(err => console.log(err));
    });

    return {
      statusCode: 200,
      body: `Success! Added subscriptions for the following channels: ${addedIdsString}`
    };
  } catch (error) {
    return { statusCode: 500, body: error.toString() };
  }
};
