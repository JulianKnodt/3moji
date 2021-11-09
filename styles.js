import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
export const styles = StyleSheet.create({
  wrapper: {
    width:"100%",
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
  },
    container: {
      flex: 1,
      backgroundColor: '#fff',
      alignItems: 'center',
      justifyContent: 'center',
      padding: 20
    },
    input: {
      height: 40,
      width: 180,
      margin: 12,
      borderWidth: 1,
      padding: 10,
    },
    inviteContainer:{
      // flex:1,
      // width:"100%",
      alignItems: 'center',
      justifyContent: 'space-around',
      display: 'flex',
      padding: 25,
    },
    inviteInput:{
      width: 65,
      height:35,
      padding:5,
      backgroundColor: "#2196F3",
      marginBottom:30,
      marginLeft:5,
    },
    reactContainer:{
      flex: 1,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'center',
    },
    inviteText:{
      fontSize: 20,
    },
    inviteButton:{
      width: 75,
      height: 75,
      padding: 5,
    },
    button: {
      width: 200,
      // borderRadius: 200,
      padding: 10,
    },
    fatButton: {
      width: '90%',
      padding: 5,
    },
    friendList:{
      width: '100%',
      borderBottomColor: 'grey',
      borderBottomEndRadius: 1,
    },
    addFriendContainer:{
      padding: 10,
      width: '50%',
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    baseText: {
      fontFamily: "Cochin"
    },
    titleText: {
      fontSize: 20,
      fontWeight: "bold",
      textDecorationLine: 'underline'
    },
    subtitleText:{
      fontWeight: 'bold'
    },
    TOSTextView:{
      width: "100%"
    },
    stretch: {
      width: 250,
      height: 50,
      resizeMode: 'stretch',
    },
    mainContent: {
      // flex: 1,
      width: 400,
      // height: "100%"
      // alignItems: 'center',
      // justifyContent: 'space-between',
    },
    hiddenInput:{
      height: 0,
      display: 'none'
    },
    centeredView: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      marginTop: 22
    },
    modalView: {
      margin: 20,
      backgroundColor: "white",
      borderRadius: 20,
      padding: 35,
      alignItems: "center",
      shadowColor: "#000",
      shadowOffset: {
        width: 0,
        height: 2
      },
      shadowOpacity: 0.25,
      shadowRadius: 4,
      elevation: 5
    },
    buttonAgree: {
      backgroundColor: "#2196F3",
      margin: 10
    },
    buttonDisagree: {
      backgroundColor: "#F194FF",
      margin: 10
    },
    textStyle: {
      color: "white",
      fontWeight: "bold",
      textAlign: "center"
    },
  });
  
