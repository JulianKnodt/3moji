import { StatusBar } from 'expo-status-bar';
import React from 'react';
import {views} from '../constants';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
import {styles} from '../styles';
import { gotoViewTest } from './utils';
export const Splash = ({stack,setStack,setCurrentView}) => {
    console.log(stack)
    return <View style={styles.container}>
      <Text>📭📩🙌!</Text>
      <View style={styles.button}>
        <Button title="Sign In" onPress={gotoViewTest(stack,views.SignIn,views.Splash,setStack,setCurrentView)}/>
      </View>

      <View style={styles.button}>
        {/* <Button title="Sign Up" onPress={() => gotoViewTest(stack,views.SignUp,views.Splash,setStack,setCurrentView)}/> */}
      </View>

      <StatusBar style="auto"/>
    </View>
};
