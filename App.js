import { StatusBar } from 'expo-status-bar';
import React, { Component } from 'react';
import { StyleSheet, Text, View, Button } from 'react-native';

// what
const views = [
  "splash",
  "sign_up",
  "sign_in",
  "home",
  "send_msg",
  "recv_msg",
];

class MainApp extends Component {
  state = {view: 0}
  set_state_sign_up() {
    this.setState({view: 1})
  }
  set_state_sign_in() {
    this.setState({view: 2})
  }
  render() {
    const v = views[this.state.view];
    if (v == "splash")
      return splash(this.set_state_sign_in.bind(this), this.set_state_sign_up.bind(this));
    else if (v == "sign_up") return sign_up();
    else if (v == "sign_in") return sign_in();
    else throw `Unknown view {v}`;
  }
}

const splash = (sign_in, sign_up) => (
  <View style={styles.container}>
    <Text>📭</Text>
    <Button title="Sign In" onPress={sign_in}>Sign In</Button>
    <Button title="Sign Up" onPress={sign_up}>Sign Up</Button>
    <StatusBar style="auto" />
  </View>
);

const sign_up = () => {
  throw "Not implemented Error"
}

const sign_in = () => {
  throw "Not implemented Error"
}

export default MainApp;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
  },
});
