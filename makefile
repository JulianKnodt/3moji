#https://stackoverflow.com/questions/33172857/how-do-i-force-a-subtree-push-to-overwrite-remote-changes
push_server:
	#git subtree push --prefix server heroku master
	git push heroku `git subtree split --prefix server master`:master --force
