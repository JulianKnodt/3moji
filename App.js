import { StatusBar } from 'expo-status-bar';
import React, { Component, useState } from 'react';
import { StyleSheet, Text, TextInput, View, Button } from 'react-native';

// what views are available for the app
const views = {
  Splash: "Splash",
  SignUp: "SignUp",
  SignIn: "SignIn",
  Home: "Home",
  SendMsg: "SendMsg",
  RecvMsg: "RecvMsg",
}

class MainApp extends Component {
  // react state
  state = {view: views.Splash}
  // stack of which views were visited
  stack = [];
  // current views state (in theory can put this on the stack too)

  goto_view(view) {
    this.stack.push(this.state);
    this.setState({view: view});
  }
  clear_stack() { this.stack = [] };
  back() {
    const prev = this.stack.pop();
    if (prev!=undefined) this.setState(prev);
  }
  async login() {
    // TODO fill this in with a server address and actually use it
    // const resp = await fetch("http://localhost:8080");
    // const json = await resp.json();
    const tempUser = {
      name: "juju",
      email: "jknodt@princeton.edu",
    }
    this.setState({user: tempUser});
    this.goto_view(views.Home)
  }
  async signup() {
    throw "NotImplementedError"
  }
  validate_email() {
    const error = (() => {
      if (!this.state.email) return null
      if (this.state.email == "") return null;
      if (!this.state.email.endsWith("princeton.edu")) {
        return "Only accepting princeton emails currently.";
      };
    })();
    this.setState({email_error: error});
  }
  render() {
    // TODO when you restart the app you always get sent to the splash screen how to do you make
    // it so that you get sent to the home screen?
    const login = this.login.bind(this);
    const signup = this.signup.bind(this)
    const back = this.back.bind(this);

    const v = views[this.state.view];
    if (v == views.Splash)
      return splash(() => this.goto_view(views.SignIn), () => this.goto_view(views.SignUp));
    else if (v == views.SignUp) return sign_up(
      back,
      email => { this.state.email = email; this.validate_email(); },
      name => { this.state.name = name },
      () => { this.signup() },
      this.state.email_error,
    );
    else if (v == views.SignIn) return sign_in(
      back,
      () => {
        login().catch(err => alert("Something went wrong 😱!\n" + err))
      },
      this.state.email_error,
    );
    else if (v == views.Home) return home();
    else throw `Unknown view {v}`;
  }
}

const splash = (sign_in, sign_up) => (
  <View style={styles.container}>
    <Text>📭📩🙌!</Text>
    <Button title="Sign In" onPress={sign_in}/>
    <Button title="Sign Up" onPress={sign_up}/>
    <StatusBar style="auto"/>
  </View>
);

// component for signing up for the app
const sign_up = (back, set_email, set_name, done, email_error) => (
  <View style={styles.container}>
    <Text>Please fill in your email:</Text>
    <TextInput
      style={styles.input}
      keyboardType="email-address"
      autoCapitalize="none"
      placeholder="@princeton.edu"
      onChangeText={set_email}
    />
    {email_error && <Text>{email_error}</Text>}
    <Text>And name:</Text>
    <TextInput
      style={styles.input}
      autoCapitalize="none"
      placeholder="Hi, my name is: 🥸"
      onChangeText={set_name}
    />
    <Button title="Sign Up" onPress={sign_up}/>
    <Button title="Back" onPress={back}/>
  </View>
);

// component for signing in to the app.
const sign_in = (back, login, email_error) => (
  <View style={styles.container}>
    <Text>Please fill in your Princeton Email:</Text>
    <TextInput
      style={styles.input}
      keyboardType="email-address"
      placeholder="@princeton.edu"
      autoCapitalize="none"
      onSubmitEditting={login}
    />
    {email_error && <Text>{email_error}</Text>}
    <Button title="Login" onPress={() => { login() }}/>
    <Button title="Back" onPress={back}/>
  </View>
);

const home = () => <View style={styles.container}>
  <Text>TODO add navigation between sending and receiving messages</Text>
</View>;

const send_msg = (friends) => <View style={styles.container}>
  {friends.map(friend => (
    <>
      <Text>{friend.name}</Text>
      <Button title="Send Msg"/>
    </>
  ))}
</View>

const draft_msg = send_msg => <View>
  <TextInput
    style={styles.input}
    // TODO automatically bring up emoji picker
    // keyboardType="email-address"
    placeholder="Emojis :)"
    // TODO only accept exactly 3 emojis
    onChange={validate_emoji}
  />
  <Button title="Send" onPress={send_msg}/>
</View>;

const ack_msg = ack => <View>
  <Button title="Thumbs Up" onPress={ack}/>
  <Button title="Hourglass" onPress={ack}/>
  <Button title="Thumbs Down" onPress={ack}/>
</View>

export default MainApp;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
  },
   input: {
    height: 40,
    width: 180,
    margin: 12,
    borderWidth: 1,
    padding: 10,
  },
});
