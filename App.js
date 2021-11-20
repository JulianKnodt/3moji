import { StatusBar } from 'expo-status-bar';
import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable, ScrollView,Modal, Image } from 'react-native';
import { Header,Tab, TabView } from 'react-native-elements';
import EmojiBoard from 'react-native-emoji-board';
import { views, HeaderText } from './constants'
import { styles } from './styles';

import * as Crypto from 'expo-crypto';
import * as Queries from './queries';
import * as Location from 'expo-location';

import AsyncStorage from '@react-native-async-storage/async-storage';

const loginTokenKey = "@3moji-login-token";
const userKey = "@3moji-user";
const emojiRegex = /(\u00a9|\u00ae|[\u2000-\u3300]|\ud83c[\ud000-\udfff]|\ud83d[\ud000-\udfff]|\ud83e[\ud000-\udfff])/;

const displayEmoji = emojis => {
  const dashs = ['-','-','-'];
  return emojis + dashs.slice([...emojis].length).join(" ");
};

const isEmail = email => {
  const re = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  return re.test(email);
}


const getLoc = async() =>{
  try{
    const { status } = await Location.requestForegroundPermissionsAsync();
    if (status !== 'granted') return null;
    const loc = await Location.getCurrentPositionAsync({accuracy:3});
    return Location.reverseGeocodeAsync(loc.coords);
  } catch(e){
    if(e.message == "Location provider is unavailable. Make sure that location services are enabled."){
      getLoc();
    }
  }
}

// TODO compute emoji len accurately
const emojiLen = (emojis) => {
  throw "NotImplementedException";
};


const MainApp = () => {
  const [user,setUser] = useState({});
  const [friends, setFriends] = useState([]);
  const [groups, setGroups] = useState([]);
  const [joinedGroups, setJoinedGroups] = useState([]);
  const [notJoinedGroups, setNotJoinedGroups] = useState([]);
  const [invites, setInvites] = useState([]);
  const [messaging, setMessaging] = useState({});
  const [viewingGroup, setViewingGroup] = useState({});
  const [stack, setStack] = useState([]);

  const [password, setPassword] = useState("");
  const [users, setUsers] = useState([]);
  const [loginToken, setLoginToken] = useState(null);
  const [currentView,setCurrentView] = useState(views.Splash);
  const [location, setLocation] = useState(null);
  const [errorMsg, setErrorMsg] = useState(null);

  useEffect(() => {
    setLocation(getLoc());
  }, []);


  useEffect(() => {
    loadLoginToken().then(token => {
      if (token == null || new Date(Number(token.validUntil) * 1000) < Date.now()) return
      setLoginToken(token);
      setCurrentView(views.Home);
      Queries.recvMsg(token).then(resp => {
        console.log("recvMsg",resp);
      });
    });
  }, []);
  const CommonHeader = props => {
    return <View style={styles.wrapper}>
      <Header centerComponent={{text: HeaderText[props.currentView]}}/>
      {props.children}
    </View>
    };

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
    (async () => {
      const pushNotifError = await Queries.registerForPushNotifications(loginToken);
      if (pushNotifError !== null) alert(pushNotifError.msg);
      console.log(pushNotifError)
    })()
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
      if (!isEmail(email)) return "Does not appear to be an email";
      return null;
    })();
    setEmailError(error);
  }

  // component for signing up for the app
  const SignUp = () => {
    const [email,setEmail] = useState("");
    const [name,setName] = useState("");
    const [password, setPassword] = useState("");
    const [emailError,setEmailError] = useState("");
    const [modalVisible, setModalVisible] = useState(false);
    const consentFormLink = "https://docs.google.com/document/d/e/2PACX-1vQRRPw63heYKNfKMYzUIY4e4txQJnexTcWoADN8cmH5Tz2a5rQ7vCeYPbSx91K5YyxrtEV81yfgNER4/pub"
    return <View style={styles.container}>
      <Modal
        animationType="slide"
        transparent={true}
        visible={modalVisible}
        onRequestClose={() => {
          Alert.alert("Modal has been closed.");
          setModalVisible(!modalVisible);
        }}
      >
      <View style={styles.centeredView}>
          <View style={styles.modalView}>
          {TOSText}
            <Pressable
              style={[styles.button, styles.buttonAgree]}
              onPress={() => signup(name, email, password).catch(err => alert("Something went wrong 😱!\n" + err))}
            >
              <Text style={styles.textStyle}>Agree</Text>
            </Pressable>
            <Pressable
              style={[styles.button, styles.buttonDisagree]}
              onPress={() => setModalVisible(!modalVisible)}
            >
              <Text style={styles.textStyle}>Disagree</Text>
            </Pressable>
          </View>
        </View>
      </Modal>
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
        <Button title="Sign Up" onPress={async () =>{
        setModalVisible(true);
      }
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
          <Button title="✍️❓" onPress={() => gotoView(views.SendMsg)}/>
        </View>
        <View style={styles.button}>
          <Button title="📨❗👀" onPress={() => gotoView(views.RecvMsg)}/>
        </View>
        <View style={styles.button}>
          <Button title="➕👥🥰" onPress={() => gotoView(views.AddGroup)}/>
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
      <Text></Text>
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

  const AckMsg = () => {
    const [emojis, setEmoji] = useState("");
    const [sendEmoji, setSendEmoji] = useState("");
    const [emojiError, setEmojiError] = useState("");
    const [messages, setMessages] = useState([]);
    const [replies, setReplies] = useState([]);
    const [sents, setSents] = useState([]);
    const [message, setMessage] = useState({});
    const [messageIndex, setMessageIndex] = useState(-1);
    useEffect(() => {
      if (emojiError === "") {
        replyMessage(message,sendEmoji).then(()=>getMessages());
        // setMessages(messages.filter((v, i) => i != messageIndex));
        setSendEmoji("")
        setMessageIndex(-1)
      } else {
        setSendEmoji("");
        setMessageIndex(-1)
      }
    }, [sendEmoji,messageIndex,message]);
    const onEnterText = (emoji,message,i) => {
      if(!emojiRegex.test(emoji)){
        setEmojiError("You can only send emojis");
      }else {
        setSendEmoji(emoji);
        setMessageIndex(i);
        setMessage(message)
        setEmojiError("")
      }
    }
    const getMessages = () => {
      Queries.recvMsg(loginToken).then(resp => {
        if(resp == null){
          setMessages([]);
          setReplies([]);
        }else{
          if(resp.newMessages == null){
            setMessages([]);
            setSendEmoji("");
            setEmojiError("");
          }else{
            const received = resp.newMessages.filter(nw => nw.source.email != loginToken.userEmail);
            setMessages(received);
            setSendEmoji("")
            setEmojiError("")
          }
          if(resp.newReplies == null){
            setReplies([]);
            setSents([]);
          }else{
            setSents(resp.newReplies.filter(nr => nr.from.email == loginToken.userEmail));
            setReplies(resp.newReplies.filter(nr => nr.message.source.email == loginToken.userEmail));
          }
        }
      });
    }
    useEffect(() => {
      getMessages();
    }, []);
    const replyMessage = async(message,reply) => {
      // console.log("reply",reply)
      const resp = await Queries.ackMsg(message.uuid,reply,loginToken);
      // console.log("reply resp",resp);
    }
    const [index,setIndex] = React.useState(0)
    useEffect(()=>{
      if(index < 0){
        setIndex(0);
      }
    },[index])
    return <View style={styles.container}>
      <Tab value={index} onChange={setIndex}>
        <Tab.Item title="📬‼️" />
        <Tab.Item title="✉️↩️"/>
        <Tab.Item title="💬🤘"/>
      </Tab>
      <TabView value={index-1} onChange={setIndex} >
        <TabView.Item styles={styles.mainContent}>
          <View styles={styles.mainContent}>
          {emojiError !== "" && <Text>{emojiError}</Text>}
          <ScrollView showsVerticalScrollIndicator={true} persistentScrollbar={true} contentContainerStyle={styles.mainContent}>
            {messages.map((message,i)=>(
                <View key={i} style={styles.inviteContainer}>
                  <Text style={styles.inviteText}>{message.source.name}📲{message.sentTo}: {message.emojis}?</Text>
                  <Text>{message.location}</Text>
                  <View style={styles.reactContainer}>
                    <View style={styles.inviteButton}>
                      <Button title="👍" onPress={()=>{
                        setSendEmoji("👍");
                        setMessageIndex(i);
                        setMessage(message);
                        }}/>
                    </View>
                    <View style={styles.inviteButton}>
                      <Button title="👎" onPress={()=>{
                        setSendEmoji("👎");
                        setMessageIndex(i);
                        setMessage(message)
                        }}/>
                    </View>
                    <TextInput
                      style={styles.inviteInput}
                      textAlign={'center'}
                      onChangeText={(text)=>{onEnterText(text,message,i)}}
                      placeholder={"➕"}
                      value={""}
                      />
                  </View>
                </View>
            ))}
          </ScrollView>
          </View>
        </TabView.Item>
        <TabView.Item styles={styles.mainContent}>
        <View>
          <ScrollView showsVerticalScrollIndicator={true} persistentScrollbar={true} contentContainerStyle={styles.mainContent}>
            {sents.map((reply,i)=>(
              <View key={i} style={styles.inviteContainer}>
              <Text style={styles.inviteText}>{reply.message.source.name}📲{reply.message.sentTo}: {reply.message.emojis}?</Text>
              <Text>{reply.message.location}</Text>
              <Text style={styles.inviteText}>{reply.reply}</Text>
              </View>
            ))}
          </ScrollView>
          </View>
        </TabView.Item>
        <TabView.Item styles={styles.mainContent}>
          <View>
          <ScrollView showsVerticalScrollIndicator={true} persistentScrollbar={true} contentContainerStyle={styles.mainContent}>
            {replies.map((reply,i)=>(
              <View key={i} style={styles.inviteContainer}>
              <Text style={styles.inviteText}>{reply.message.source.name}📲{reply.message.sentTo}: {reply.message.emojis}?</Text>
              <Text>{reply.message.location}</Text>
              <Text style={styles.inviteText}>{reply.from.name}:{reply.reply}</Text>
              </View>
            ))}
            </ScrollView>
          </View>
        </TabView.Item>
      </TabView>
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    {/* <EmojiBoard showBoard={show} onClick={(emoji)=>{replyMessage(message,uuid,emoji.code)}}/> */}
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
    return <View style={styles.container}>
      {notJoinedGroups.map(group => (
        <View key={group.uuid} style={styles.button}>
          <Button
            title={group.name}
            onPress={()=>{
              setViewingGroup(group);
              gotoView(views.ViewGroup);
            }}
          />
        </View>

      ))}
      <View style={styles.button}>
        <Button title="🆕👥📨" onPress={()=>{gotoView(views.CreateGroup)}}/>
      </View>
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
    </View>
  };
  const ViewGroup = ({viewingGroup}) => {
    // console.log("viewing",viewingGroup)
    return <View style={styles.container}>
      <Text>{viewingGroup.name}</Text>
      <Text>Members:{Object.values(viewingGroup.users).join(",")}</Text>
      <View style={styles.button}>
        <Button title="Join"
          onPress={async ()=>{
              const resp = await Queries.joinGroup(loginToken,viewingGroup.uuid);
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
  }
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
  if (currentView == views.Splash) return <Splash gotoView={gotoView.bind(this)}/>;
  else if (currentView == views.SignUp) return <SignUp/>;
  else if (currentView == views.SignIn) return <SignIn/>;
  else if (currentView == views.Home) return <CommonHeader currentView={currentView}><Home/></CommonHeader>;
  else if (currentView == views.SendMsg) return <CommonHeader currentView={currentView}><SendMsg/></CommonHeader>;
  else if (currentView == views.RecvMsg) return <CommonHeader currentView={currentView}><AckMsg/></CommonHeader>;
  else if (currentView == views.DraftMsg) return <CommonHeader currentView={currentView}><DraftMsg
    messaging={messaging}
    gotoView={gotoView}
    getGroups={getGroups}
    loginToken={loginToken}
    back={back}
  /></CommonHeader>;
  else if (currentView == views.AddFriend) return <AddFriend/>;
  else if (currentView == views.AddGroup) return <CommonHeader currentView={currentView}><AddGroup/></CommonHeader>;
  else if (currentView == views.CreateGroup) return <CommonHeader currentView={currentView}><CreateGroup/></CommonHeader>;
  else if (currentView == views.ViewGroup) return <CommonHeader currentView={currentView}><ViewGroup viewingGroup={viewingGroup}/></CommonHeader>
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

const Splash = props => {
  return <View style={styles.container}>
    <Text>📭📩🙌!</Text>
    <View style={styles.button}>
      <Button title="Sign In" onPress={() => props.gotoView(views.SignIn)}/>
    </View>
    <View style={styles.button}>
      <Button title="Sign Up" onPress={() => props.gotoView(views.SignUp)}/>
    </View>
    <StatusBar style="auto"/>
  </View>
};

const DraftMsg = props => {
  const { messaging, getGroups, gotoView, back, loginToken } = props;
  const [emojis, setEmoji] = useState("");
  const [emojiError, setEmojiError] = useState("");
  const [loc, setLoc] = useState("");
  const [show, setShow] = useState(false);
  const [recommendations, setRecommendations] = useState([]);

  useEffect(() => {
    const fetchMessage = async() =>{
      const resp = await Queries.recommendations();
      // console.log(resp.recommendations);
      setRecommendations(resp.recommendations || []);
    }
    fetchMessage();
  }, []);

  const sendEmoji = async() => {
    // console.log("emojis",emojis)
    if (emojis.length != 6) return setEmojiError("You need to send exactly three emojis");
    const resp = await Queries.sendMsg(loginToken, emojis, messaging.uuid, loc);
    if (resp instanceof Queries.Error) {
      alert(resp.msg);
    } else back();
  }

  const onEnterText = emoji => {
    if(emoji.length < emojis.length){
      setEmoji(emoji);
      return;
    }
    const newText = emoji.substring(emojis.length);
    if (emojis.length >= 6) setEmojiError("You can only add three emojis");
    else if(!emojiRegex.test(newText)){
      // console.log(newText)
      setEmojiError("You can only send emojis");
    }
    else {
      setEmoji(emoji);
      setEmojiError("");
    }
  };

  const onRemove = () => {
    if (emojis.length > 0) setEmoji(emojis.substring(0, emojis.length - 2));
    else if(emojis.length <= 6) setEmojiError("");
  }
  // console.log(emojis)
  return <View style={styles.container}>
    <Text>Sending message to {messaging.name}</Text>
    <Text>Members:{Object.values(messaging.users).join(",")}</Text>
    <TextInput
        style={styles.input}
        onChangeText={onEnterText}
        value={emojis}
        placeholder="✍️😀❓"
    />
    {/* <Pressable onPress={() => setShow(!show)}>
        <Text>{displayEmoji(emojis)}</Text>
    </Pressable> */}
    {/* <EmojiBoard showBoard={show} onClick={onClick} onRemove={onRemove} /> */}
    {emojiError !== "" && <Text>{emojiError}</Text>}
    <View style={styles.button}>
      <Button title="Send" onPress={sendEmoji}/>
    </View>
    <View style={styles.button}>
      {loc ? <Text>{loc}</Text> : null}
      <Button title={`${loc ? "➖" : "➕"}🗺📍`} onPress={async () => {
        if(loc){
          setLoc("")
          return;
        }
        const l = await getLoc();
        console.log(l);
        if (!l || l.length == 0) return alert("Could not get location!")
        const locString = l[0].street ? `${l[0].name}, ${l[0].street}` : `${l[0].name}`;
        setLoc(locString);
      }}/>
    </View>
    {recommendations.map((recommendation,i)=>(
      <View key={i} style={styles.button}>
        <Button title={recommendation} color="#5ac18e" onPress={()=>{setEmoji(recommendation)}}></Button>
      </View>
    ))}
    <View style={styles.button}>
      <Button title="Leave Group" color="#b81010" onPress={async () => {
        const resp = await Queries.leaveGroup(loginToken, messaging.uuid, loc);
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


export const TOSText = (
  <ScrollView showsVerticalScrollIndicator={true} persistentScrollbar={true} contentContainerStyle={styles.TOSTextView}>
    <Image style={styles.stretch} source={require('./consent_form.png')}/>
    <Text>
      TITLE OF RESEARCH: Effect of Limited (emoji) choice, Better defaults, & Ephemeral messages on communication
      {"\n\n"}
      PRINCIPAL INVESTIGATOR: Monroy-Hernández, Andrés
      {"\n\n"}
      PRINCIPAL INVESTIGATOR’S DEPARTMENT: Computer Science
    </Text>
    <Text style={styles.titleText}>
      Key information about the study:
    </Text>
    <Text>
      Your informed consent is being sought for research.
      Participation in the research is voluntary.
      The purpose of the research is to understand how users interacted with broadcasted ephemeral emoji messages.
      {"\n\n"}
      The procedures that the subject will be asked to follow in the research: Messages and reaction data will be gathered for analysis(anonymized).
      An optional survey will be released after about a month of App release.
      Both components of the study are optional and voluntary.
      {"\n\n"}
      The reasonably foreseeable risks or discomforts to the subject as a result of participation: Minimal
      {"\n\n"}
      The benefits to the subject or to others, e.g., society that may reasonably be expected from the research : Understand the role of ephemerality and emojis in communication.
    </Text>
    <Text style={styles.titleText}>
      Additional information about the study:
    </Text>
    <Text style={styles.subtitleText}>
    Confidentiality:
    </Text>
    <Text>
    All records from this study will remain anonymous. Your responses will be kept private, and we will not include any information that will make it possible to identify you in any report we might publish.
    {"\n\n"}
    Research records will be stored securely in a locked cabinet and/or on password-protected computers. The research team will be the only party that will have access to your data.
    </Text>
    <Text style={styles.subtitleText}>
    Who to contact with questions:
    </Text>
    <Text>
    Investigators: yl1128@princeton.edu or jknodt@princeton.edu
    {"\n\n"}
    If you have questions regarding your rights as a research subject, or if problems arise which you do not feel you can discuss with the Investigator, please contact the Institutional Review Board at: Phone: (609) 258-8543 Email: irb@princeton.edu
    </Text>
    <Text style={styles.titleText}>
    Summary:
    </Text>
    <Text>
    I understand the information that was presented and that:
    {"\n\n"}
    My participation is voluntary.
    {"\n\n"}
    Refusal to participate will involve no penalty or loss of benefits to which I am otherwise entitled. I may discontinue participation at any time without penalty or loss of benefits.
    {"\n\n"}
    I do not waive any legal rights or release Princeton University or its agents from liability for negligence. I hereby give my consent to be the subject of the research.

    </Text>
  </ScrollView>
)
