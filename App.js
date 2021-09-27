import { StatusBar } from 'expo-status-bar';
import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
import EmojiBoard from 'react-native-emoji-board';
import { views } from './constants'
import { styles } from './styles';
import * as Crypto from 'expo-crypto';
import AsyncStorage from '@react-native-async-storage/async-storage';

const serverURL = "https://api-3moji.herokuapp.com/";
const loginTokenKey = "@3moji-login-token";

const displayEmoji = emojis => {
  const dashs = ['-','-','-'];
  return emojis + dashs.slice(emojis.length).join(" ");
};

const saveLoginToken = async (token) => {
  try {
    if (!token) return await AsyncStorage.removeItem(loginTokenKey);
    await AsyncStorage.setItem(loginTokenKey, JSON.stringify(token));
  } catch (e) {
    // saving error
    console.log("failed", e)
  }
};

const friendActions = {
  onlyFriends: 0,
  all: 1,
  notFriends: 2,
};
// common HTTP headers.
const headers = {
  Accept: 'application/json', 'Content-Type': 'application/json'
};

const getPeople = async (loginToken, amount=50, kind=friendActions.onlyFriends) => {
    const body = { kind, amount, loginToken };
    const resp = await fetch(serverURL + "api/v1/people/", {
      method: 'POST', headers,
      body: JSON.stringify(body),
    });
    if (resp.status != 200) {
      console.log(resp.status);
    };
    return resp.json();
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
}

const MainApp = () => {
  const [user,setUser] = useState({});
  const [friends, setFriends] = useState([]);
  const [invites, setInvites] = useState([]);
  const [messaging, setMessaging] = useState({});
  const [stack, setStack] = useState([]);
  const [show, setShow] = useState(false);

  const [password, setPassword] = useState("");
  const [users, setUsers] = useState([]);
  const [loginToken, setLoginToken] = useState(null);
  const [currentView,setCurrentView] = useState(views.Splash);

  useEffect(() => {
    loadLoginToken().then(token => {
      if (token == null) return
      setLoginToken(token);
      setCurrentView(views.Home);
    });
  }, []);

  const updateFriendsAndInvites = async () => {
    const friends = await getPeople(loginToken);
    console.log(friends);
  };

  // when a login token is acquired, will reload friends list and get current invitations.
  useEffect(() => {
    if (loginToken == null) return;
    updateFriendsAndInvites();
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
    if (prev!=undefined) setCurrentView(prev);
  }
  const successEntry = respJSON => {
    saveLoginToken(respJSON.loginToken);
    setLoginToken(respJSON.loginToken);
    setUser(respJSON.user);

    gotoView(views.Home)
  };

  const login = async (email, password) => {
    const digest = await Crypto.digestStringAsync(
      Crypto.CryptoDigestAlgorithm.SHA256,
      password,
    );
    const resp = await fetch(serverURL + "api/v1/login/", {
      method: 'POST',
      headers,
      body: JSON.stringify({
        email,
        hashedPassword: digest,
      }),
    });
    if (resp.status !== 200) {
      console.log(resp.status);
    } else successEntry(await resp.json());
  }
  const signup = async (name, email, password) =>{
    const digest = await Crypto.digestStringAsync(
      Crypto.CryptoDigestAlgorithm.SHA256,
      password,
    );
    console.log(name, email, digest);
    const dst = serverURL + "api/v1/sign_up/";
    const resp = await fetch(dst, {
      method: 'POST',
      headers,
      body: JSON.stringify({ email, name, hashedPassword: digest }),
    });
    if (resp.status !== 200) {
      console.log(resp.status);
    } else successEntry(await resp.json());
  }
  const  validateEmail = (email,setEmailError) => {
    const error = (() => {
      console.log(email)
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
        onChangeText={(text) => {
          setName(text);
        }
      }
      />

      <Text>{"and password:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        placeholder="my password is: 🔐 "
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
      <Text>{"Please fill in your Princeton Email:"}</Text>
      <TextInput
        style={styles.input}
        keyboardType="email-address"
        placeholder="@princeton.edu"
        autoCapitalize="none"
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
          onPress={() => gotoView(views.SendMsg)}
        />

      </View>
      <View style={styles.button}>
        <Button
          title="📫😆❗"
          onPress={() => gotoView(views.RecvMsg)}
        />
      </View>
      <View style={styles.button}>
        <Button
          title="➕😊🥰"
          onPress={() => gotoView(views.AddFriend)}
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
      {friends.map(friend => (
        <>
          <View style={styles.button}>
            <Button
              title={friend.name}
              onPress={()=>{
                setMessaging(friend);
                gotoView(views.DraftMsg);
            }}/>
          </View>

        </>
      ))}
      {/* </View> */}
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };

  const sendEmoji = () => {
    if (emojis.length != 6) {
      setEmojiError("You need to send exactly three emojis");
      return
    }
    const req = {
      loginToken: loginToken,
    }
    fetch(serverURL + "api/v1/send_msg/", {
      method: "POST",
      headers,
      body: JSON.stringify({
        loginToken,
        message: {
          emojis: emojis,
          source: user,
          // TODO fill this in with the recipient.
          recipients: [],
          sentAt: Date.now(),
        },
      }),
    });
  }

  const DraftMsg = () => {
    const [emojis, setEmoji] = useState("");
    const [emojiError, setEmojiError] = useState("");

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
      if(emojis.length < 6){
        setEmojiError("");
      }
    }
    return <View style={styles.container}>
    <Text>Sending message to {messaging.name}</Text>
    <Pressable onPress={() => setShow(!show)}>
        <Text>{displayEmoji(emojis)}</Text>
    </Pressable>
    <EmojiBoard showBoard={show} 
    onClick={onClick} onRemove={onRemove}
    />
    {emojiError !== "" && <Text>{emojiError}</Text>}
    <View style={styles.button}>
      <Button title="Send" onPress={sendEmoji}/>
    </View>
    <View style={styles.button}>
      <Button title="Back" color="#f194ff" onPress={back}/>
    </View>
  </View>};

  const AckMsg = () => {
    const [emojis, setEmoji] = useState("");
    const [emojiError, setEmojiError] = useState("");
    return <View style={styles.container}>
      {/* <View style={styles.mainContent}> */}
    {invites.map(invite=>(
        <View  style={styles.inviteContainer}>
          <Text style={styles.inviteText}>{invite.from.name}: {invite.emojis}?</Text>
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

  if (currentView == views.Splash){
    return <Splash />;
  } else if (currentView == views.SignUp){
    return <SignUp />;
  } else if (currentView == views.SignIn){
    return <SignIn />;
  }
  else if (currentView == views.Home) return <Home />;
  else if (currentView == views.SendMsg) return <SendMsg />;
  else if (currentView == views.RecvMsg) return <AckMsg />;
  else if (currentView == views.DraftMsg) return <DraftMsg />;
  else if (currentView == views.AddFriend) return <AddFriend />;
  else throw `Unknown view {currentView}`;

}

export default MainApp;
