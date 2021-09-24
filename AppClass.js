// class MainApp extends Component {
//   // react state
//   state = {view: views.Splash}
//   // stack of which views were visited
//   stack = [];
//   // current views state (in theory can put this on the stack too)

//   goto_view(view) {
//     this.stack.push(this.state);
//     this.setState({view: view});
//   }
//   clear_stack() { this.stack = [] };
//   back() {
//     const prev = this.stack.pop();
//     if (prev!=undefined) this.setState(prev);
//   }
//   async login() {
//     // TODO fill this in with a server address and actually use it
//     // const resp = await fetch("http://localhost:8080");
//     // const json = await resp.json();
//     const tempUser = {
//       name: "juju",
//       email: "jknodt@princeton.edu",
//     }
//     this.setState({
//                   user: tempUser,
//                   friends: [{name:'YX',email:"yx.edu"},{name:'Chen',email:'qc.edu'}],
//                   invites: [{from:{name:'YX',email:"yx.edu"},emojis:"🍫🍦🍰"},
//                             {from:{name:'Chen',email:'qc.edu'},emojis:"🍣🍜🍛"}],
//                 });
//     this.goto_view(views.Home)
//   }
//   async signup() {
//     throw "NotImplementedError"
//   }
//   validate_email() {
//     const error = (() => {
//       if (!this.state.email) return null
//       if (this.state.email == "") return null;
//       if (!this.state.email.endsWith("princeton.edu")) {
//         return "Only accepting princeton emails currently.";
//       };
//     })();
//     this.setState({email_error: error});
//   }
//   render() {
//     // TODO when you restart the app you always get sent to the splash screen how to do you make
//     // it so that you get sent to the home screen?
//     const login = this.login.bind(this);
//     const signup = this.signup.bind(this)
//     const back = this.back.bind(this);
//     const goto_view = this.goto_view.bind(this);
//     const set_state = this.setState.bind(this);

//     const v = views[this.state.view];
//     if (v == views.Splash)
//       return splash(() => this.goto_view(views.SignIn), () => this.goto_view(views.SignUp));
//     else if (v == views.SignUp) return sign_up(
//       back,
//       email => { this.state.email = email; this.validate_email(); },
//       name => { this.state.name = name },
//       () => { this.signup() },
//       this.state.email_error,
//     );
//     else if (v == views.SignIn) return sign_in(
//       back,
//       () => {
//         login().catch(err => alert("Something went wrong 😱!\n" + err))
//       },
//       this.state.email_error,
//     );
//     else if (v == views.Home) return home(goto_view,back);
//     else if (v == views.SendMsg) return send_msg(this.state.friends,set_state,goto_view,back);
//     else if (v == views.RecvMsg) return ack_msg(this.state.invites,back);
//     else if (v == views.DraftMsg) return draft_msg(back,this.state.messaging);
//     else if (v == views.AddFriend) return add_friend(back);
//     else throw `Unknown view {v}`;
//   }
// }

// const splash = (sign_in, sign_up) => (
//   <View style={styles.container}>
//     <Text>📭📩🙌!</Text>
//     <View style={styles.button}>
//       <Button title="Sign In" onPress={sign_in}/>
//     </View>
    
//     <View style={styles.button}>
//       <Button title="Sign Up" onPress={sign_up}/>
//     </View>
    
//     <StatusBar style="auto"/>
//   </View>
// );

// // component for signing up for the app
// const sign_up = (back, set_email, set_name, done, email_error) => (
//   <View style={styles.container}>
//     <Text>Please fill in your email:</Text>
//     <TextInput
//       style={styles.input}
//       keyboardType="email-address"
//       autoCapitalize="none"
//       placeholder="@princeton.edu"
//       onChangeText={set_email}
//     />
//     {email_error && <Text>{email_error}</Text>}
//     <Text>And name:</Text>
//     <TextInput
//       style={styles.input}
//       autoCapitalize="none"
//       placeholder="Hi, my name is: 🥸"
//       onChangeText={set_name}
//     />
//     <View style={styles.button}>
//       <Button title="Sign Up" onPress={sign_up}/>
//     </View>
    
//     <View style={styles.button}>
//       <Button title="Back" color="#f194ff" onPress={back}/>
//     </View>
    
//   </View>
// );

// // component for signing in to the app.
// const sign_in = (back, login, email_error) => (
//   <View style={styles.container}>
//     <Text>Please fill in your Princeton Email:</Text>
//     <TextInput
//       style={styles.input}
//       keyboardType="email-address"
//       placeholder="@princeton.edu"
//       autoCapitalize="none"
//       onSubmitEditting={login}
//     />
//     {email_error && <Text>{email_error}</Text>}
//     <View style={styles.button}>
//       <Button title="Login" onPress={() => { login() }}/>
//     </View>
    
//     <View style={styles.button}>
//       <Button title="Back" color="#f194ff" onPress={back}/>
//     </View>
    
//   </View>
// );

// const home = (goto_view,back) => <View style={styles.container}>
//   <View style={styles.button}>
//     <Button
//       title="✉️🥺❓"
//       onPress={() => goto_view(views.SendMsg)}
//     />
    
//   </View>
//   <View style={styles.button}>
//     <Button 
//       title="📫😆❗" 
//       onPress={() => goto_view(views.RecvMsg)}
//     />
//   </View>
//   <View style={styles.button}>
//     <Button 
//       title="➕😊🥰"
//       onPress={() => goto_view(views.AddFriend)}
//     />
//   </View>
//   <View style={styles.button}>
//     <Button title="Back" color="#f194ff" onPress={back}/>
//   </View>
// </View>;

// const send_msg = (friends,set_state,goto_view,back) => <View style={styles.container}>
//   {console.log(friends)}
//   {friends.map(friend => (
//     <>
//     {/* <View style={styles.friendList}>
//       <Pressable onPress={()=>{}}>  
//         <Text>{friend.name}</Text>
//       </Pressable>
//     </View> */}
//       <View style={styles.button}>
//         <Button
//           title={friend.name} 
//           onPress={()=>{
//             set_state({messaging:friend});
//             goto_view(views.DraftMsg);
//         }}/>
//       </View>
      
//     </>
//   ))}
//   <View style={styles.button}>
//     <Button title="Back" color="#f194ff" onPress={back}/>
//   </View>
// </View>

// const draft_msg = (back,messaging) => <View style={styles.container}>
//   {/* <TextInput
//     style={styles.input}
//     // TODO automatically bring up emoji picker
//     // keyboardType="email-address"
//     placeholder="Emojis :)"
//     // TODO only accept exactly 3 emojis
//     // onChange={validate_emoji}
//   /> */}
//   <Pressable onPress={() => {}}>
//       <Text>click here</Text>
//   </Pressable>
//   <EmojiBoard showBoard={true} onClick={(emoji)=>{console.log(emoji)}} />
//   {/* <EmojiInput
// 	onEmojiSelected={(emoji) => {console.log(emoji)}}
// 	/> */}
//   <View style={styles.button}>
//     <Button title="Send" onPress={()=>{}}/>
//   </View>
//   <View style={styles.button}>
//     <Button title="Back" color="#f194ff" onPress={back}/>
//   </View>
// </View>;

// const ack_msg = (invites,back) => <View style={styles.container}>
//   {invites.map(invite=>(
//       <View  style={styles.inviteContainer}>
//         <Text style={styles.inviteText}>{invite.from.name}: {invite.emojis}?</Text>
//         <View style={styles.reactContainer}>
//           <View style={styles.inviteButton}>
//             <Button title="👍" onPress={()=>{}}/>
//           </View>
//           <View style={styles.inviteButton}>
//             <Button title="👎" onPress={()=>{}}/>
//           </View>
//           <View style={styles.inviteButton}>
//             <Button title="⌛" onPress={()=>{}}/>
//           </View>
//       </View>
//     </View>
//   ))}
//   <View style={styles.button}>
//     <Button title="Back" color="#f194ff" onPress={back}/>
//   </View>
// </View>;

// const add_friend = (back) => <View style={styles.container}>
//   <Text>Please enter your friend's email address:</Text>
//   <TextInput
//     style={styles.input}
//     keyboardType="email-address"
//     autoCapitalize="none"
//     placeholder="@princeton.edu"
//     onChangeText={()=>{}}
//   />
//   <View style={styles.button}>
//     <Button title="➕🤗💛" onPress={()=>{}}/>
//   </View>
//   <View style={styles.button}>
//     <Button title="Back" color="#f194ff" onPress={back}/>
//   </View>
// </View>;