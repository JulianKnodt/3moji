import React, { Component, useState, useEffect } from 'react';
import { StyleSheet, Text, TextInput, View, Button, Pressable } from 'react-native';
export const styles = StyleSheet.create({
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
      width: "100%",
      height: "20%",
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: 25,
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
      width: 50,
      height: 50,
      padding: 10,
    },
    button: {
      width: '50%',
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
    mainContent: {
      flex: 1,
      alignItems: 'center',
      justifyContent: 'center',
    }
  });
  
