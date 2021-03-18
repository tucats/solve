


# Table of Contents
1. [Introduction](#intro)
2. [Server Commands](#commands)
    1. [Starting and Stopping](#startstop)
    2. [Credentials Management](#credentials)
    3. [Profile Settings](#profile)
3. [Writing a Service](#services)
    1. [Global Variables](#globals)
    2. [Server Directives](#directives)
    3. [Server Functions](#functions)
    4. [Sample Service](#sample)


&nbsp;
&nbsp;

# Ego Web Server <a name="intro"></a>

This documents using the _Ego_ web server capability. You can start Ego as a REST server,
with a specified port on which to listen for input (the default is 8080). Web service
requests are handled by the server for administrative functions like logging in or 
managing user credentials, and by _Ego_ programs for other service functions that
represent the actual web services features.
&nbsp;
&nbsp;

## Server subcommands <a nanme="commands"></a>
The `ego server` command has subcommands that describe the operations you can perform. The
commands that start or stop a rest server or evaluate it's status must be run on the same
computer that the server itself is running on. For each of the commands below, you can 
specify the option `--port n` to indicate that you want to control the server listening 
on the given port number, where `n` is an integer value for a publically available port 
number.

| Subcommand | Description |
|------------| ------------|
| start | Start a server. You can start multiple servers as long as they each have a different --port number assigned to them. |
| stop | Stop the server that is listening on the named port. If the port is not specified, then the default port is assumed. |
| restart | Stop the current server and restart it with the exact same command line values. This can be used to restart a server that has run out of memory, or when upgrading the version of ego being used. |
| status | Report on the status of the server. |
| users set | Create or update a user in the server database |
| users delete | Remove a user from the server database |
| users list | List users in the server database |
| caches list | List the endpoints currently in the service cache |
| caches flush | Flush the service cache on the server |
| caches set-size | Set the number of service endpoints the cache can hold |

&nbsp;
&nbsp;

The commands that start and stop a server only required native operating system permissions
to start or stop a process. The commands that affect user credentials in the server can only 
be executed when logged into the server with a process that has `root` privileges, as defined 
by the credentials database in that server.

When a server is running, it generates a log file (in the current directory, by default) which 
tracks the server startup and status of requests made to the server.

### Starting and Stopping the Server<a name="startstop"></a>

The `ego server start` command accepts command line options to describe the port on which to 
listen,  whether or not to use secure HTTPS, and options that control how authentication is 
handled. The  `ego server stop` command stops a running server. The `ego server restart` stops 
and restarts a server using the options it was used to start up originally.

You can also run the server from the CLI (instead of detaching it as a process) using the
`ego server run` command option, which accepts the same options as `ego server start` and runs
the code directly in the CLI, sending logging to stdout.

When a server is started, a file is created (by default in /tmp) that describes the server
status and command-line options. This information is re-read when issuing a `ego server status`
command to display server information. It is also read by the `ego server restart` command to
determine the command-line options to use with the restarted server.

When a server is stopped via `ego server stop`, the server status file in /tmp is deleted.

Below is additional information about the options that can be used for the `start` and `run`
commands.

#### Caching
You can specify a cache size, which controls how many service programs are held in memory and not
recompiled. This can be a significant performance benefit. When an endpoint call is made, the
server checks to see if the cache already contains the compiled code for that function along with
it's package definitions. If so, it is reused to execute the current service. 

If the service
program was not in the cache, it will be added to the cache.  When the cache becomes full (has
met the limit on the number of programs to cache) then the least-recently-used service program
based on timestamp of the last REST call) is removed from the cache. 

**NOTE: THIS FEATURE IS CURRENTLY NOT WORKING CORRECTLY. DO NOT SPECIFY A CACHE SIZE.** 
The default cache size is currently set to zero, which disables caching entirely.

#### /code Endpoint
By default, the server will only run services already stored in the services directory tree
(more on that below). When you start the web service, you can optionally enable the `\code`
endpoint. This accepts a text body and runs it as a program directly. This can be used for
debugging purposes or diagnosing issues with a server. This should **NOT** be left enabled
by default, as it exposes the server to security risks.

#### Logging
By default, the server generates a lot file named "ego-server.log" in the current directory
where the `server start` command is issued. This contains entries describing server operations
(like records of endpoints called, and HTTP status returned). It also contains a periodic
display of memory consumption by the server.

You can override the location of the log file using the `--log` command line option, and
specifying the location and file name where the log file is to be written. The log will
continue to be written to as long as the server is running. Note that the first line of
the log file contains the UUID of the server session, so you can correlate a log to a
running instance of the server.

You can specify `--no-log` if you wish to suppress logging.

#### Port and Security
Specify the `--port` option to indicate the integer port number that _Ego_ should use to
listen for REST requests. If not specified, the default is port 8080. You can have multiple
_Ego_ servers running at one time, as long as they each use a different port number. The
port number is also used in other commands like `server status` to report on the status of
a particular instance of the server on the current computer.

By default, _Ego_ servers assume HTTPS communication. This requires that you have specified
a suitable trusted certificate store in the default location on your system that _Ego_ can
use to verify server trust.

If you wish to run in insecure mode, you can use the "--not-secure" option direct the server 
to listen for HTTP requests that are not encrypted. You must not use this mode in a 
production environment, since it is possible for users to snoop for username/password 
pairs and authentication tokens. It  is useful to debug issues where you are attempting 
to isolate whether your are having an issue with trust certificates or not.

#### Authentication
An _Ego_ web server can serve endpoints that require authentication or not, and whether
the authentication is done by username/password versus an authentication token. Server
command options control where the authentication credentails are stored, the default 
"super user" account, and the security "realm" used for password challenges to web
clients.

* Use the `--users` command line option to specify either the file system path and
  file name to use for local JSON data that contains the credentials information, or
  a "postgres://" URL expression that indicates the Postgres database used to store
  the credentials (in a schema named "ego-server" that is created if needed).
* Use the `--superuser` option to specify a "username:password" string indicating 
  the default superuser. This is only needed when the credentials store is first
  initialized; it creates a user with the given username and password and gives that
  user the "ROOT" privilege which makes them able to perform all secured operations.
* Use the "--realm" option to specify a string that is sent back to web clients when
  a username/password is required but was not provided. For web clients that are
  browsers, this string is typically displayed in the username/password prompt from
  the browser.
  
&nbsp;
&nbsp;

## Credentials Management <a name="credentials"></a>

Use the `ego logon` command to logon to the server you have started, using a username and
password that has root/admin privileges. This communicates with the web server and asks it
to issue a token that is used for all subsequent administraiton operations. This token is
valid for 24 hours by default; after 24 hours you must log in again using the username and
password.

Once you have logged in, you can issue additional `ego server` commands to manage the 
credentials database used by the web server, and manage the service cache used to
reduce re-compilation times for services used frequently.



&nbsp;
&nbsp;

## Profile items <a name="profile"></a>
The REST server can be easily controlled by persistent items in the current profile,
which are set with the `ego profile set` command or via program operation using the
`profile` package.

| item | description |
|------| ------------|
| ego.logon.defaultuser | A string value of "user:pass" describing the default credential to apply when there is no user database |
| ego.logon.userdata | the path to the JSON file containing the user data |
| ego.token.expiration | the default duration a token is considered value. The default is "15m" for 15 minutes |
| ego.token.key | A string used to encrypt tokens. This can be any string value |
&nbsp; 
&nbsp;

# Writing a Service <a name="services"></a>

This section covers details of writing a service. The service program is called automatically
by the Ego web server when a request comes in with an endpoint URL that matches the service
file location. The URL is available to the service, along with other variables indicating
status of authentication, etc. The service program is then run, and it has responsibility
for determining the HTTP status and response type of the result.


Server startup scans the `services/` directory below the Ego path to find the Ego programs
that offer endpoint support. This directory structure will map to the endpoints that the
server responds to.  For example, a service program named `foo` in the `services/` directory 
will be referenced with an endoint like http://host:port/services/foo

It is the responsibility of each endpoint to do whatever validation is requireed for
the endpoint. To help support this, a number of global variables are set up for the
endpoint service program which describe  information about the rest call and the
credentials (if any) of the caller.

## Global Variables <a name="#globals"></a>
Each time a REST call is made, the program associated with the endpoint is run. When it
runs, it will have a number of global variables set already that the program can use
to control its operation.

| Name        | Type    | Description                                         |
|-------------|---------|-----------------------------------------------------|
| _body       | string  | The text of the body of the request.                |
| _headers    | struct  | A struct where the field is the header name, and the value is an array of string values for each value found  |
| _parms      | struct  | A struct where the field name is the parameter, and the value si an array of string values for each value found |
| _password   | string  | The Basic authentication password provided, or empty string if not used |
| _superuser  | bool | The username or token identity has root privileges |
| _token      | string  | The Token authentication value, or an empty string if not used |
| _token_valid | bool | The Token authentication value is a valid token |
| _user       | string  | The Basic authentication username provided, or identify of a valid Token |
&nbsp; 
&nbsp;     

## Server Directives <a name="#directives"></a>
There are a few compiler directives that can be used in service programs that are executed 
by the server. These allow for more declarative code.

`@authenticated type`
This requires that the caller of the service be authenticated, and specifies the type of the 
authentication to be performed. This should be at the start of the service code; if the caller 
is not authenticated then the rest of the services does not run.  Valid types are:

| Type | Description |
| --- | --- |
| any | User can be authenticated by username or token |
| token | User must be authenticated by token only |
| user | User must be authenticated with username/password only |
| admin | The user (regardless of authentication) must have root privileges |
| tokenadmin | The user must authenticated by token and have root privilieges |

&nbsp;
&nbsp;

`@status n`
This sets the REST return status code to the given integer value. By default, the status value i
s 200 for "success" but can be set to any valid integer HTTP status code. For example, 404 means 
"not found" and 403 means "forbidden".

`@response v`
This adds the contents of the expression value `v` to the result set returned to the caller. You
can have multiple `@response` directives, they are accumulated in the order executed. A primary 
value of this is also that it automatically detects if the media type is meant to specify JSON
results; in this case the value is automatically converted to a JSON string before being added
to the response.

## Functions <a name="functions"></a>
There are additional functions made available to the Ego programs run as services. These are 
generally used to support writing services for administrative or privileged functions. For example, 
a service that updates a password probably would use all of the following functions.

| Function | Description | 
|----------|-------------|
| u := getuser(name) | Get the user data for a given user
| call setuser(u) | Update or create a user with the given user data
| f := authenticated(user,pass) | Boolean if the username and password are valid

&nbsp; 
&nbsp;     

## Sample Service <a name="sample"></a>
This section describes the source for a simple service.