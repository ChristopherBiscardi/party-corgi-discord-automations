const fetch = require('node-fetch');

exports.handler = async (event, context) => {
  try {
    const streamerMarkdown = await fetch(
      'https://raw.githubusercontent.com/ChristopherBiscardi/party-corgi/master/docs/list-of-streamers.md'
    )
      .then(res => res.text())
      .then(body => {
        return body;
      });

    const mixerRegex = new RegExp(
      /(https?:\/\/(.+?\.)?mixer\.com(\/[A-Za-z0-9\-\._~:\/\?#\[\]@!$&'\(\)\*\+,;\=]*)?)/g
    );
    const mixerUrls = Array.from(streamerMarkdown.match(mixerRegex));

    const twitchRegex = new RegExp(
      /(https?:\/\/(.+?\.)?twitch\.tv(\/[A-Za-z0-9\-\._~:\/\?#\[\]@!$&'\(\)\*\+,;\=]*)?)/g
    );
    const twitchUrls = Array.from(streamerMarkdown.match(twitchRegex));

    const response = {
      mixer: mixerUrls,
      twitch: twitchUrls
    };
    console.log(response);

    return {
      statusCode: 200,
      body: JSON.stringify(response)
    };
  } catch (err) {
    return { statusCode: 500, body: `ERROR: ${err.toString()}` };
  }
};
