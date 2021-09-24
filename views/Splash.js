import { StatusBar } from 'expo-status-bar';
import React from 'react';
import {views} from '../constants';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
import {styles} from '../styles';
export const Splash = (gotoView) => {
    console.log(gotoView)
    return <View style={styles.container}>
      <Text>📭📩🙌!</Text>
      <View style={styles.button}>
        <Button title="Sign In" onPress={gotoView(views.SignIn)}/>
      </View>
      
      <View style={styles.button}>
        <Button title="Sign Up" onPress={() => gotoView(views.SignUp)}/>
      </View>
      
      <StatusBar style="auto"/>
    </View>
};