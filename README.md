# Blog-Aggregator

## Requirements

go 1.24+
PostgressSQl 16+

## Install

go install https://github.com/aa3447/Blog-Aggregator

## Usage

Manually create a config file in your home directory, ~/.gatorconfig.json, with the following content:
{"db_url":"protocol://username:password@host:port/database","current_user_name":""}

Login sets the current user in the config if the user exists in the database.
Requires one argument: username.
Blog-Aggregator login jake_the_snake

Register creates a new user in the database and sets it as the current user in the config.
Requires one argument: username.
Blog-Aggregator register jake_the_snake

Users retrieves and prints all users from the database, indicating the current user.
Blog-Aggregator users

Feeds retrieves and prints all feeds from the database, including their creators.
Blog-Aggregator feeds

Agg fetches and prints items from feeds at regular intervals.
Requires one argument: time between requests.
Blog-Aggregator agg 1(s,m,h)

AddFeed creates a new feed in the database and automatically follows it for the current user.
Requires two arguments: feed name and feed URL.
user must be logged in.
Blog-Aggregator addfeed Boot.dev https://blog.boot.dev/index.xml

Follow creates a new feed follow relationship in the database for the current user and the specified feed URL.
Requires one argument: feed URL.
user must be logged in.
Blog-Aggregator follow https://blog.boot.dev/index.xml

Following lists all feeds followed by the current user.
user must be logged in.
Blog-Aggregator following

Unfollow removes a feed follow relationship in the database for the current user and the specified feed URL.
Requires one argument: feed URL.
Blog-Aggregator unfollow https://blog.boot.dev/index.xml

Browse retrieves and prints posts for the current user with pagination.
Requires two arguments: limit and offset.
user must be logged in.
Blog-Aggregator Browse 10 0

Reset removes all users from the database and clears the current user in the config.
Blog-Aggregator reset

gator
