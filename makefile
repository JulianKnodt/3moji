#https://stackoverflow.com/questions/33172857/how-do-i-force-a-subtree-push-to-overwrite-remote-changes
push_server:
	#git subtree push --prefix server heroku master
	git push heroku `git subtree split --prefix server master`:master --force

# really dumb but I don't know how to reset the server otherwise and can't SSH in.
reset_redis:
	heroku addons:destroy redis --confirm api-3moji
	heroku addons:create redis
