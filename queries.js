import * as Crypto from 'expo-crypto';
import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';

const serverURL = "https://api-3moji.herokuapp.com/";
const headers = {
  Accept: 'application/json', 'Content-Type': 'application/json',
};

export class Error {
  constructor(status=400, msg="") {
    this.status = status;
    this.msg = msg;
  }
}

export const listPeopleKind = {
  onlyFriends: 0,
  all: 1,
  notFriends: 2,
}

export const getPeople = async (loginToken, amount=50, kind=listPeopleKind.onlyFriends) => {
  const req = { amount, kind, loginToken };
  console.log(req);
  const resp = await fetch(serverURL + "api/v1/list_friends/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

export const listGroupKind = {
  allGroups: 0,
  joinedGroups: 1,
  notJoinedGroups: 2,
};

export const getGroups = async (loginToken, amount=50, kind=listGroupKind.allGroups) => {
  const req = { amount, kind, loginToken };
  const resp = await fetch(serverURL + "api/v1/list_groups/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

export const groupOpKind = {
  joinGroup: 0,
  leaveGroup: 1,
  createGroup: 2,
};

export const joinGroup = async (loginToken, groupUuid) =>
  groupOp(loginToken, "", groupUuid, groupOpKind.joinGroup);

export const leaveGroup = async (loginToken, groupUuid) =>
  groupOp(loginToken, "", groupUuid, groupOpKind.leaveGroup);

export const createGroup = async (loginToken, groupName) =>
  groupOp(loginToken, groupName, null, groupOpKind.createGroup);

const groupOp = async (
  loginToken, groupName="", groupUuid=null, kind=groupOpKind.joinGroup,
) => {
  console.log(groupName,groupUuid);
  if (kind == groupOpKind.joinGroup || kind == groupOpKind.leaveGroup) {
    // requires a groupUuid
    if (!groupUuid) return null;
  } else if (kind == groupOpKind.createGroup) {
    // requires a groupName
    if (groupName.length == 0) return null;
  }
  const req = { loginToken, kind, groupName, groupUuid };
  const resp = await fetch(serverURL + "api/v1/groups/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp,true);
};

const localTime = () => {
  const now = new Date();
  const localTime = now.getHours() + now.getMinutes()/60 + now.getSeconds()/3600;
  return localTime;
};

export const sendMsg = async (loginToken, emojis, dstUuid, loc="", toGroup=true) => {
  const recipientKind = toGroup ? 0 : 1;
  // TODO message is not just a string but a complex object.
  const message = {
    uuid: loginToken.uuid,
    emojis: emojis,
    // source will be populated on the server.
    location: loc,
    sentAt: Date.now().toString(),
    localTime: localTime().toString(),
  };
  const req = { message, loginToken, recipientKind, to: dstUuid };
  const resp = await fetch(serverURL + "api/v1/send_msg/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp, true);
};

export const recvMsg = async (loginToken) => {
  const req = { loginToken, deleteOld: false, };
  const resp = await fetch(serverURL + "api/v1/recv_msg/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
}

export const ackMsg = async (msgID, reply, loginToken) => {
  const req = { msgID, reply, loginToken };
  console.log("reply",msgID,reply);
  if(reply === "") return;
  const resp = await fetch(serverURL + "api/v1/ack_msg/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp,true);
};

export const recommendations = async () => {
  const req = { localTime: localTime() };
  const resp = await fetch(serverURL + "api/v1/recs/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

export const signup = async (name, email, password) => {
  const digest = await Crypto.digestStringAsync(
    Crypto.CryptoDigestAlgorithm.SHA256, password,
  );
  const req = { email, name, hashedPassword: digest };
  const dst = serverURL + "api/v1/sign_up/";
  const resp = await fetch(dst, {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

export const login = async (email, password) => {
  const digest = await Crypto.digestStringAsync(
    Crypto.CryptoDigestAlgorithm.SHA256, password,
  );
  const req = { email, hashedPassword: digest };
  const dst = serverURL + "api/v1/login/";
  const resp = await fetch(dst, {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

export const notifTokenActionKind = {
  add: 0,
  rm: 1,
};

const granted = "granted";
export const registerForPushNotifications = async loginToken => {
  if (!Constants.isDevice) return alert('Must use physical device for Push Notifications');

  const existingStatus = (await Notifications.getPermissionsAsync()).status;
  let finalStatus = existingStatus;
  if (existingStatus !== granted)
    finalStatus = (await Notifications.requestPermissionsAsync()).status;

  if (finalStatus !== granted) return alert('Failed to get push token for push notification!');
  const token = (await Notifications.getExpoPushTokenAsync()).data;
  // just re-get token each time?

  if (Platform.OS === 'android') {
    Notifications.setNotificationChannelAsync('default', {
      name: 'default',
      importance: Notifications.AndroidImportance.MAX,
      vibrationPattern: [0, 250, 250, 250],
      lightColor: '#FF231F7C',
    });
  }
  const req = { token, loginToken, kind: notifTokenActionKind.add, };
  const dst = serverURL + "api/v1/push_token/";
  const resp = await fetch(dst, {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp, true);
};

export const unsubPushNotifications = async loginToken => {
  const req = { token, loginToken, kind: notifTokenActionKind.rm, };
  const dst = serverURL + "api/v1/push_token/";
  const resp = await fetch(dst, {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp, true);
};

// current generic way to handle responses, returning null if there's an error which may be
// turned into an alert.
const handleResp = async (resp,ignoreResp = false) => {
  if (resp.status >= 300) return new Error(resp.status, await resp.text());

  if (ignoreResp) return null;
  return resp.json();
};

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: false,
    shouldSetBadge: false,
  }),
});
