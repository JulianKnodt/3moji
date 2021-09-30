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
  groupOp(loginToken, groupName, 0, groupOpKind.leaveGroup);

const groupOp = async (
  loginToken, groupName="", groupUuid=0, kind=groupOpKind.joinGroup,
) => {
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
  return handleResp(resp);
};

export const sendMsg = async (loginToken, message, dstUuid, toGroup=true) => {
  recipientKind = toGroup ? 0 : 1;
  // TODO message is not just a string but a complex object.
  const req = { message, loginToken, recipientKind, to: dstUuid };
  const resp = await fetch(serverURL + "api/v1/send_msg/", {
    method: 'POST', headers, body: JSON.stringify(req),
  });
  return handleResp(resp);
};

// current generic way to handle responses, returning null if there's an error which may be
// turned into an alert.
const handleResp = async resp => {
  if (resp.status != 200) {
    return new Error(resp.status, await resp.text());
  }
  return resp.json();
};
