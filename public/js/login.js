// FirebaseUI config.
var uiConfig = {
    callbacks: {
        signInSuccessWithAuthResult: (authResult, redirectUrl) => {
            console.log(authResult);
            console.log(redirectUrl);
            return false;
        }
    },
    // signInSuccessUrl: '<url-to-redirect-to-on-success>',
    signInOptions: [
        // Leave the lines as is for the providers you want to offer your users.
        firebase.auth.GoogleAuthProvider.PROVIDER_ID,
        // firebase.auth.FacebookAuthProvider.PROVIDER_ID,
        // firebase.auth.TwitterAuthProvider.PROVIDER_ID,
        // firebase.auth.GithubAuthProvider.PROVIDER_ID,
        firebase.auth.EmailAuthProvider.PROVIDER_ID,
        // firebase.auth.PhoneAuthProvider.PROVIDER_ID,
        // firebaseui.auth.AnonymousAuthProvider.PROVIDER_ID
    ],
    // // tosUrl and privacyPolicyUrl accept either url string or a callback
    // // function.
    // // Terms of service url/callback.
    // tosUrl: '<your-tos-url>',
    // // Privacy policy url/callback.
    // privacyPolicyUrl: function () {
    //     window.location.assign('<your-privacy-policy-url>');
    // }
};


let postIdTokenToSessionLogin = (url, idToken) => {
    const req = new Request(url, {method: "POST", body: idToken});
    return fetch(req);
};
// As httpOnly cookies are to be used, do not persist any state client side.
firebase.auth().setPersistence(firebase.auth.Auth.Persistence.NONE);


// Initialize the FirebaseUI Widget using Firebase.
var ui = new firebaseui.auth.AuthUI(firebase.auth());


let handleSignedInUser = function(user) {
    console.log(user);
    return user.getIdToken().then(idToken => {
        console.log(idToken)
        return postIdTokenToSessionLogin('/sessionLogin', idToken);
    }).then(() => {
        return firebase.auth().signOut();
    }).then(() => {
        window.location.assign('/');
    })
};

let handleSignedOutUser = function() {
    // The start method will wait until the DOM is loaded.
    ui.start('#firebaseui-auth-container', uiConfig);
};

firebase.auth().onAuthStateChanged(function(user) {
    user ? handleSignedInUser(user) : handleSignedOutUser();
})