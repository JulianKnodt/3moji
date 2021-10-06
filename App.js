import { StatusBar } from 'expo-status-bar';
import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
import EmojiBoard from 'react-native-emoji-board';
import { views } from './constants'
import { styles } from './styles';
import * as Crypto from 'expo-crypto';
import * as Queries from './queries';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Location from 'expo-location';

const loginTokenKey = "@3moji-login-token";
const userKey = "@3moji-user";

const displayEmoji = emojis => {
  const dashs = ['-','-','-'];
  return emojis + dashs.slice(emojis.length).join(" ");
};

const MainApp = () => {
  const [user,setUser] = useState({});
  const [friends, setFriends] = useState([]);
  const [groups, setGroups] = useState([]);
  const [joinedGroups, setJoinedGroups] = useState([]);
  const [notJoinedGroups, setNotJoinedGroups] = useState([]);
  const [invites, setInvites] = useState([]);
  const [messaging, setMessaging] = useState({});
  const [stack, setStack] = useState([]);
  const [show, setShow] = useState(false);

  const [password, setPassword] = useState("");
  const [users, setUsers] = useState([]);
  const [loginToken, setLoginToken] = useState(null);
  const [currentView,setCurrentView] = useState(views.Splash);
  const [location, setLocation] = useState(null);
  const [errorMsg, setErrorMsg] = useState(null);

  const getLocation = async() =>{
    try{
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        setErrorMsg('Permission to access location was denied');
        return;
      }

      let location = await Location.getCurrentPositionAsync({accuracy:3});
      const geo = await Location.reverseGeocodeAsync(location.coords);
      setLocation(geo);
    } catch(e){
      if(e.message == "Location provider is unavailable. Make sure that location services are enabled."){
        getLocation();
      }
    }
  }
  useEffect(() => {
    getLocation();
  }, []);

  useEffect(() => {
    loadLoginToken().then(token => {
      if (token == null) return
      setLoginToken(token);
      setCurrentView(views.Home);
    });
  }, []);

  const updateFriendsAndInvites = async () => {
    const friends = await Queries.getPeople(loginToken,50,Queries.listPeopleKind.all);
    if (friends instanceof Queries.Error) {
      return alert(friends.msg);
    } else setFriends(friends);
    // TODO fetch invites
    setInvites([{name:"A group",message:"🥞🍳🥓"}]);
  };

  const getGroups = async () => {
    const group = await Queries.getGroups(loginToken);
    if (group instanceof Queries.Error) {
      alert(group.msg);
    } else {
      if (group == null || group.groups == null) {
        setGroups([]);
      } else setGroups(group.groups);
    }
    const joined = await Queries.getGroups(loginToken,50,Queries.listGroupKind.joinedGroups);
    if (joined instanceof Queries.Error) {
      alert(joined.msg);
    } else {
      if (joined == null || joined.groups == null) {
        setJoinedGroups([]);
      } else setJoinedGroups(joined.groups);
    };
    const notJoined = await Queries.getGroups(loginToken,50,Queries.listGroupKind.notJoinedGroups);
    if (joined instanceof Queries.Error) {
      alert(joined.msg);
    } else{
      if (notJoined == null || notJoined.groups == null) {
        setNotJoinedGroups([]);
      } else setNotJoinedGroups(notJoined.groups);
    };
  }

  // when a login token is acquired, will reload friends list and get current invitations.
  useEffect(() => {
    if (loginToken == null) return;
    updateFriendsAndInvites();
    getGroups();
  }, [loginToken]);

  // TODO fetch friends and invites
  const gotoView = (view) => {
    setStack([...stack,currentView]);
    setCurrentView(view);
  }

  const clearStack = () => {
    setStack([]);
  }

  const back = () => {
    const prev = stack.pop();
    if (prev != undefined) setCurrentView(prev);
  }
  const successEntry = respJSON => {
    saveLoginToken(respJSON.loginToken);
    setLoginToken(respJSON.loginToken);
    setUser(respJSON.user);

    gotoView(views.Home)
  };

  const login = async (email, password) => {
    const resp = await Queries.login(email, password);
    if (resp instanceof Queries.Error) {
      alert(resp.msg);
    } else successEntry(resp);
  }
  const signup = async (name, email, password) => {
    const resp = await Queries.signup(name, email, password);
    if (resp instanceof Queries.Error) {
      alert(resp.msg);
    } else successEntry(resp);
  };
  const  validateEmail = (email,setEmailError) => {
    const error = (() => {
      if (!email) return null
      if (email == "") return null;
      if (!email.endsWith("princeton.edu")) {
        return "Only accepting princeton emails currently.";
      };
    })();
    setEmailError(error);
  }

  const Splash = () => {
    return <View style={styles.container}>
      <Text>📭📩🙌!</Text>
      <View style={styles.button}>
        <Button title="Sign In" onPress={() => gotoView(views.SignIn)}/>
      </View>
      <View style={styles.button}>
        <Button title="Sign Up" onPress={() => gotoView(views.SignUp)}/>
      </View>
      <StatusBar style="auto"/>
    </View>

  };

  // component for signing up for the app
  const SignUp = () => {
    const [email,setEmail] = useState("");
    const [name,setName] = useState("");
    const [password, setPassword] = useState("");
    const [emailError,setEmailError] = useState("");
    return <View style={styles.container}>
      <Text>{"Please fill in your email:"}</Text>
      <TextInput
        style={styles.input}
        keyboardType="email-address"
        autoCapitalize="none"
        placeholder="@princeton.edu"
        defaultValue={email}
        onChangeText={(text) => {
            setEmail(text);
            validateEmail(text,setEmailError);
          }
        }
      />
      {emailError !== "" && <Text>{emailError}</Text>}
      <Text>{"username:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        placeholder="Hi, my name is: 🥸"
        onChangeText={text => setName(text)}
      />

      <Text>{"and password:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        placeholder="my password is: 🔐"
        secureTextEntry={true}
        onChangeText={setPassword}
      />
      <View style={styles.button}>
        <Button title="Sign Up" onPress={async () =>
        signup(name, email, password).catch(err => alert("Something went wrong 😱!\n" + err))
        }/>
      </View>

      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={() => back()}/>
      </View>
    </View>
  };

  // component for signing in to the app.
  const SignIn = () => {
    const [email,setEmail] = useState("");
    const [name,setName] = useState("");
    const [password, setPassword] = useState("");
    const [emailError,setEmailError] = useState("");
    return <View style={styles.container}>
      <Text>{"Please fill in your email:"}</Text>
      <TextInput
        style={styles.input}
        keyboardType="email-address"
        placeholder="@princeton.edu"
        autoCapitalize="none"
        autoComplete="off"
        autoCorrect={false}
        onChangeText={(text) => {
          setEmail(text);
          validateEmail(text,setEmailError);
        }
      }
      />
      {emailError !== "" && <Text>{emailError}</Text>}
      <Text>{"and password:"}</Text>
      <TextInput
        style={styles.input}
        placeholder="🔐"
        autoCapitalize="none"
        secureTextEntry={true}
        onChangeText={setPassword}
      />
      <View style={styles.button}>
        <Button title="Login" onPress={async () => {
          login(email, password).catch(err => alert("Something went wrong 😱!\n" + err))}}/>
      </View>
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };

  const Home = () => (
    <View style={styles.container}>
      <View style={styles.button}>
        <Button
          title="✉️🥺❓"
          onPress={() => {
            gotoView(views.SendMsg)}}
        />

      </View>
      <View style={styles.button}>
        <Button
          title="📨❗👀"
          onPress={() => gotoView(views.RecvMsg)}
        />
      </View>
      <View style={styles.button}>
        <Button
          title="➕😊🥰"
          onPress={() => gotoView(views.AddGroup)}
        />
      </View>
      <View style={styles.button}>
        <Button title="Log out" color="#f194ff" onPress={() => {
            saveLoginToken(null);
            setLoginToken(null);
            clearStack();
            setCurrentView(views.Splash);
          }
        }/>
      </View>
    </View>
  );

  const SendMsg = () => {
    return <View style={styles.container}>
      {/* <View style={styles.mainContent}> */}
      {joinedGroups.map(group => (
        <View key={group.uuid} style={styles.button}>
          <Button
            title={group.name}
            onPress={()=>{
              setMessaging(group);
              gotoView(views.DraftMsg);
          }}/>
        </View>

      ))}
      {/* </View> */}
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };

  const DraftMsg = () => {
    const [emojis, setEmoji] = useState("");
    const [emojiError, setEmojiError] = useState("");

    const sendEmoji = async() => {
      if (emojis.length != 6) {
        setEmojiError("You need to send exactly three emojis");
        return;
      }
      const resp = await Queries.sendMsg(loginToken, emojis, messaging.uuid);
      if (resp instanceof Queries.Error) {
        alert(resp.msg);
      } else back();
    }

    const onClick = emoji => {
      if (emojis.length >= 6){
        setEmojiError("You can only add three emojis");
      }else{
        setEmoji(emojis + emoji.code);
        setEmojiError("");
      }
    };

    const onRemove = () => {
      if(emojis.length > 0){
        setEmoji(emojis.substring(0, emojis.length - 2));
      }
      if(emojis.length <= 6){
        setEmojiError("");
      }
    }
    return <View style={styles.container}>
    <Text>Sending message to {messaging.name}</Text>
    <Pressable onPress={() => setShow(!show)}>
        <Text>{displayEmoji(emojis)}</Text>
    </Pressable>
    <EmojiBoard showBoard={show} onClick={onClick} onRemove={onRemove} />
    {emojiError !== "" && <Text>{emojiError}</Text>}
    <View style={styles.button}>
      <Button title="Send" onPress={sendEmoji}/>
    </View>
    <View style={styles.button}>
      <Button title="Leave Group" color="#b81010" onPress={async()=>{
        const resp = await Queries.leaveGroup(loginToken, messaging.uuid);
        if (resp instanceof Queries.Error) {
          return alert(resp.msg);
        }
        getGroups();
        gotoView(views.SendMsg);
      }}/>
    </View>
    <View style={styles.button}>
      <Button title="Back" color="#f194ff" onPress={back}/>
    </View>
  </View>};
  const AckMsg = () => {
    console.log(invites);
    const [emojis, setEmoji] = useState("");
    const [emojiError, setEmojiError] = useState("");
    return <View style={styles.container}>
      {/* <View style={styles.mainContent}> */}
    {invites.map((invite,i)=>(
        <View key={i} style={styles.inviteContainer}>
          <Text style={styles.inviteText}>{invite.name}: {invite.message}?</Text>
          <View style={styles.reactContainer}>
            <View style={styles.inviteButton}>
              <Button title="👍" onPress={()=>{}}/>
            </View>
            <View style={styles.inviteButton}>
              <Button title="👎" onPress={()=>{}}/>
            </View>
            <View style={styles.inviteButton}>
              <Button title="➕" onPress={()=>{setShow(!show)}}/>
            </View>
          </View>
        </View>
    ))}
    {/* </View> */}
    <View style={styles.button}>
      <Button title="Back" color="#f194ff" onPress={back}/>
    </View>
    <EmojiBoard showBoard={show} onClick={(emoji)=>{}}/>
  </View>};

  const AddFriend = () => {
    return <View style={styles.container}>
      {/* <View style={styles.mainContent}> */}
    {users.filter(u => u.email !== user.email).map(
      u => <>
      <View style={styles.addFriendContainer}>
        <Text>{u.name}</Text>
        <View>
          <Button title="➕" onPress={()=>{}}/>
        </View>
      </View>
    </>
    )}
    {/* </View> */}
    <View style={styles.button}>
      <Button title="Back" color="#f194ff" onPress={back}/>
    </View>
  </View>};
  const AddGroup = () => {
    // console.log(groups);
    return <View style={styles.container}>
      {notJoinedGroups.map(group => (
        <View key={group.uuid} style={styles.button}>
          <Button
            title={group.name}
            onPress={async ()=>{
              console.log("group",group);
              const resp = await Queries.joinGroup(loginToken,group.uuid);
              if (resp instanceof Queries.Error) {
                return alert(resp.msg);
              }
              console.log(resp)
              getGroups();
              setMessaging(group);
              gotoView(views.DraftMsg);
          }}/>
        </View>

      ))}
      <View style={styles.button}>
        <Button title="🆕👨‍👩‍👧‍👦📨" onPress={()=>{gotoView(views.CreateGroup)}}/>
      </View>
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };
  const CreateGroup = () => {
    const [groupName,setGroupName] = useState("")
    return <View style={styles.container}>
      <Text>{"Please enter a group name:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        value={groupName}
        onChangeText={setGroupName}
      />
      <View style={styles.button}>
        <Button title="Create" onPress={async()=>{
          const resp = await Queries.createGroup(loginToken,groupName);
          if (resp instanceof Queries.Error) {
            return alert(resp.msg);
          }
          getGroups();
          back();
        }}/>
      </View>
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };
  if (currentView == views.Splash) return <Splash/>;
  else if (currentView == views.SignUp) return <SignUp/>;
  else if (currentView == views.SignIn) return <SignIn/>;
  else if (currentView == views.Home) return <Home/>;
  else if (currentView == views.SendMsg) return <SendMsg/>;
  else if (currentView == views.RecvMsg) return <AckMsg/>;
  else if (currentView == views.DraftMsg) return <DraftMsg/>;
  else if (currentView == views.AddFriend) return <AddFriend/>;
  else if (currentView == views.AddGroup) return <AddGroup/>;
  else if (currentView == views.CreateGroup) return <CreateGroup/>;
  else throw `Unknown view {currentView}`;
};

export default MainApp;

const saveLoginToken = async (token) => {
  try {
    if (!token) return await AsyncStorage.removeItem(loginTokenKey);
    await AsyncStorage.setItem(loginTokenKey, JSON.stringify(token));
  } catch (e) {
    // saving error
    console.log("failed", e)
  }
};

const loadLoginToken = async () => {
  try {
    const loginToken = await AsyncStorage.getItem(loginTokenKey);
    if (loginToken == null) return null;
    const token = JSON.parse(loginToken);
    // TODO validate token still valid here.
    return token;
  } catch (e) {
    // retrieving error
    console.log("failed", e)
  }
};
