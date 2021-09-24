import { StatusBar } from 'expo-status-bar';
import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
import EmojiBoard from 'react-native-emoji-board';
import {views} from './constants'
// import { Splash } from './views/Splash';
import {styles} from './styles';
// what views are available for the app


const MainApp = () =>{
  const [currentView,setCurrentView] = useState(views.Splash);
  const [email,setEmail] = useState("");
  const [name,setName] = useState("");
  const [emailError,setEmailError] = useState("");
  const [user,setUser] = useState({});
  const [friends, setFriends] = useState([]);
  const [invites, setInvites] = useState([]);
  const [messaging, setMessaging] = useState({});
  const [stack, setStack] = useState([]);
  const [show, setShow] = useState(false);
  const [emojis, setEmoji] = useState("");
  const [emojiError, setEmojiError] = useState("");
  const [passWord, setPassword] = useState("");
  const [users, setUsers] = useState([]);
  // TODO actually fetch users
  useEffect(() => {  
    setUsers([
      {name: "juju", email: "jknodt@princeton.edu",},
      {name:'YX',email:"yx.edu"},
      {name:'Chen',email:'qc.edu'}
    ])
  },[]);
  const onClick = emoji => {
    console.log(emojis.length)
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
  // TODO fetch friends and invites
  console.log(stack);
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

  const login = async() => {
    // TODO fill this in with a server address and actually use it
    // const resp = await fetch("http://localhost:8080");
    // const json = await resp.json();
    const tempUser = {
      name: "juju",
      email: "jknodt@princeton.edu",
    }
    setUser(tempUser);
    setFriends([{name:'YX',email:"yx.edu"},{name:'Chen',email:'qc.edu'}]);
    setInvites([{from:{name:'YX',email:"yx.edu"},emojis:"🍫🍦🍰"},
                            {from:{name:'Chen',email:'qc.edu'},emojis:"🍣🍜🍛"}]);
    gotoView(views.Home)
  }
  const signup = async() =>{
    throw "NotImplementedError"
  }
  const  validateEmail = () => {
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
    return <View style={styles.container}>
      <Text>{"Please fill in your email:"}</Text>
      <TextInput
        style={styles.input}
        keyboardType="email-address"
        autoCapitalize="none"
        placeholder="@princeton.edu"
        onChangeText={(email) => {
            setEmail(email);
            validateEmail();
          }
        }
      />
      {emailError !== "" && <Text>{emailError}</Text>}
      <Text>{"username:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        placeholder="Hi, my name is: 🥸"
        onChangeText={setName}
      />
      
      <Text>{"and password:"}</Text>
      <TextInput
        style={styles.input}
        autoCapitalize="none"
        placeholder="my password is: 🔐 "
        onChangeText={setPassword}
      />
      <View style={styles.button}>
        <Button title="Sign Up" onPress={signup}/>
      </View>

      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={() => back()}/>
      </View>
      
    </View>
  };

  // component for signing in to the app.
  const SignIn = () => {
    return <View style={styles.container}>
      <Text>{"Please fill in your Princeton Email:"}</Text>
      <TextInput
        style={styles.input}
        keyboardType="email-address"
        placeholder="@princeton.edu"
        autoCapitalize="none"
        onChangeText={setEmail}
      />
      {emailError !== "" && <Text>{emailError}</Text>}
      <Text>{"and password:"}</Text>
      <TextInput
        style={styles.input}
        placeholder="🔐"
        autoCapitalize="none"
        onChangeText={setPassword}
      />
      <View style={styles.button}>
        <Button title="Login" onPress={() => {
          login().catch(err => alert("Something went wrong 😱!\n" + err))}}/>
      </View>
      
      <View style={styles.button}>
        <Button title="Back" color="#f194ff" onPress={back}/>
      </View>
      
    </View>
  };

  const Home = () => {return <View style={styles.container}>
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
          clearStack(); 
          setCurrentView(views.Splash)}
        }
      />
    </View>
  </View>};

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
  </View>};
  const displayEmoji = () =>{
    const dashs = ['-','-','-']
    return emojis + dashs.slice(emojis.length).join(" ");
  }

  const sendEmoji = () => {
    if(emojis.length == 6){
      // TODO actually send it
    }else{
      setEmojiError("You need to send exactly three emojis");
    }
  }

  const DraftMsg = () => { return <View style={styles.container}>
    <Text>Sending message to {messaging.name}</Text>
    <Pressable onPress={() => setShow(!show)}>
        <Text>{displayEmoji()}</Text>
    </Pressable>
    <EmojiBoard showBoard={show} onClick={onClick} onRemove={onRemove}/>
    {emojiError !== "" && <Text>{emojiError}</Text>}
    <View style={styles.button}>
      <Button title="Send" onPress={sendEmoji}/>
    </View>
    <View style={styles.button}>
      <Button title="Back" color="#f194ff" onPress={back}/>
    </View>
  </View>};

  const AckMsg = () => { 
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
    <EmojiBoard showBoard={show} onClick={onClick} onRemove={onRemove}/>
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
    return <Splash gotoView={gotoView}/>;
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

