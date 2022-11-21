package i18n

// Messages contains a map of internalized strings. The map is organized
// by language as the first key, and then the text ID passed into the
// i18n.T() function. If a key in a given language is not found, it
// reverts to using the "en" key.
var Messages = map[string]map[string]string{
	"en": {
		"ego": "run an Ego program",

		"ego.config":                 "Manage the configuration",
		"ego.config.delete":          "Delete a key from the configuration",
		"ego.config.list":            "List all configurations",
		"ego.config.remove":          "Delete an entire configuration",
		"ego.config.set":             "Set a configuration value",
		"ego.config.set.description": "Set the configuration description",
		"ego.config.set.output":      "Set the default output type (text or json)",
		"ego.config.show":            "Show the current configuration",
		"ego.logon":                  "Log onto a remote server",

		"ego.path": "Print the default ego path",

		"ego.run": "Run an existing program",

		"ego.server":                "Start to accept REST calls",
		"ego.server.caches":         "Manage server caches",
		"ego.server.cache.flush":    "Flush service caches",
		"ego.server.cache.list":     "List service caches",
		"ego.server.cache.set.size": "Set the server cache size",
		"ego.server.logging":        "Display or configure server logging",
		"ego.server.logon":          "Log on to a remote server",
		"ego.server.restart":        "Restart an existing server",
		"ego.server.run":            "Run the rest server",
		"ego.server.start":          "Start the rest server as a detached process",
		"ego.server.status":         "Display server status",
		"ego.server.stop":           "Stop the detached rest server",
		"ego.server.users":          "Manage server user database",
		"ego.server.user.set":       "Create or update user information",
		"ego.server.user.delete":    "Delete a user from the server's user database",
		"ego.server.user.list":      "List users in the server's user database",

		"ego.sql":             "Execute SQL in the database server",
		"ego.sql.file":        "Filename of SQL command text",
		"ego.sql.row-ids":     "Include the row UUID in any output",
		"ego.sql.row-numbers": "Include the row number in any output",

		"ego.table":             "Operate on database tables",
		"ego.table.create":      "Create a new table",
		"ego.table.delete":      "Delete rows from a table",
		"ego.table.drop":        "Delete one or more tables",
		"ego.table.grant":       "Set permissions for a given user and table",
		"ego.table.insert":      "Insert a row to a table",
		"ego.table.list":        "List tables",
		"ego.table.permission":  "List table permissions",
		"ego.table.permissions": "List all table permissions (requires admin privileges)",
		"ego.table.read":        "Read contents of a table",
		"ego.table.show":        "Show table metadata",
		"ego.table.sql":         "Directly execute a SQL command",
		"ego.table.update":      "Update rows to a table",
		"ego.test":              "Run a test suite",

		// Note that labels are case-sensitive, to indicate the expected case of the translation.
		"label.Active":                "Active",
		"label.command":               "command",
		"label.Commands":              "Commands",
		"label.configuration":         "configuration",
		"label.Default.configuration": "Default configuration",
		"label.Description":           "Description",
		"label.Error":                 "Error",
		"label.Field":                 "Field",
		"label.Key":                   "Key",
		"label.Logger":                "Logger",
		"label.Name":                  "Name",
		"label.options":               "options",
		"label.parameter":             "parameter",
		"label.parameters":            "parameters",
		"label.Parameters":            "Parameters",
		"label.password.prompt":       "Password: ",
		"label.Row":                   "Row",
		"label.since":                 "since",
		"label.Type":                  "Type",
		"label.Usage":                 "Usage",
		"label.Value":                 "Value",
		"label.version":               "version",

		"msg.config.written":           "Configuration key {{key}} written",
		"msg.config.deleted":           "Configuration {{name}} deleted",
		"msg.logged.in":                "Successfully logged in as {{user}}, valid until {{expires}}",
		"msg.server.cache":             "Server Cache, hostname {{host}}, ID {{id}}",
		"msg.server.cache.assets":      "There are {{count}} HTML assets in cache, for a total size of {{size}} bytes.",
		"msg.server.cache.services":    "There are {{count}} service items in cache. The maximum cache size is {{limit}} items.",
		"msg.server.cache.no.assets":   "There are no HTML assets cached.",
		"msg.server.cache.one.asset":   "There is 1 HTML asset in cache, for a total size of {{size}} bytes.",
		"msg.server.cache.one.service": "There is 1 service item in cache. The maximum cache size is {{limit}} items.",
		"msg.server.cache.no.services": "There are no service items in cache. The maximum cache size is {{limit}} items.",
		"msg.server.cache.updated":     "Server cache size updated",
		"msg.server.cache.emptied":     "Server cache emptied",
		"msg.server.logs.file":         "Server log file is {{name}}",
		"msg.server.logs.no.retain":    "Server does not retain previous log files",
		"msg.server.logs.purged":       "Purged {{count}} old log files",
		"msg.server.logs.retains":      "Server also retains last {{count}} previous log files",
		"msg.server.logs.status":       "Logging status, hostname {{host}}, ID {{id}}",
		"msg.server.not.running":       "Server not running",
		"msg.server.started":           "Server started as process {{pid}}",
		"msg.server.status":            "Ego {{version}}, pid {{pid}}, host {{host}}, session {{id}}",
		"msg.server.stopped":           "Server (pid {{pid}}) stopped",

		"opt.config.force": "Do not signal error if option not found",

		"opt.logon.server": "URL of server to authenticate with",

		"opt.sql.row.ids":     "Include the row UUID in the output",
		"opt.sql.row.numbers": "Include the row number in the output",
		"opt.sql.file":        "Filename of SQL command text",

		"opt.address.port":      "Specify address (and optionally address) of server",
		"opt.filter":            "List of optional filter clauses",
		"opt.help.text":         "Show this help text",
		"opt.insecure":          "Do not require X509 server certificate verification",
		"opt.limit":             "If specified, limit the result set to this many rows",
		"opt.password":          "Password for logon",
		"opt.port":              "Specify port number of server",
		"opt.scope":             "Blocks can access any symbol in call stack",
		"opt.start":             "If specified, start result set at this row",
		"opt.trace":             "Display trace of bytecode execution",
		"opt.username":          "Username for logon",
		"opt.symbol.allocation": "Allocation size (in symbols) when expanding storage for a symbol table ",

		"opt.global.debug":   "Debug loggers to enable",
		"opt.global.format":  "Specify text, json or indented output format",
		"opt.global.profile": "Name of profile to use",
		"opt.global.quiet":   "If specified, suppress extra messaging",
		"opt.global.version": "Show version number of command line tool",

		"opt.run.auto.import": "Override auto-import configuration setting",
		"opt.run.debug":       "Run with interactive debugger",
		"opt.run.disasm":      "Display a disassembly of the bytecode before execution",
		"opt.run.entry.point": "Name of entrypoint function (defaults to main)",
		"opt.run.log":         "Direct log output to this file instead of stdout",
		"opt.run.static":      "Enforce static typing on program execution",
		"opt.run.symbols":     "Display symbol table",

		"opt.server.delete.user":     "Username to delete",
		"opt.server.logging.enable":  "List of loggers to enable",
		"opt.server.logging.disable": "List of loggers to disable",
		"opt.server.logging.file":    "Show only the active log file name",
		"opt.server.logging.keep":    "Specify how many log files to keep",
		"opt.server.logging.session": "Limit display to log entries for this session number",
		"opt.server.logging.status":  "Display the state of each logger",
		"opt.server.run.cache":       "Number of service programs to cache in memory",
		"opt.server.run.code":        "Enable /code endpoint",
		"opt.server.run.debug":       "Service endpoint to debug",
		"opt.server.run.force":       "If set, override existing PID file",
		"opt.server.run.is.detached": "If set, server assumes it is already detached",
		"opt.server.run.keep":        "The number of log files to keep",
		"opt.server.run.log":         "File path of server log",
		"opt.server.run.no.log":      "Suppress server log",
		"opt.server.run.not.secure":  "If set, use HTTP instead of HTTPS",
		"opt.server.run.realm":       "Name of authentication realm",
		"opt.server.run.sandbox":     "File path of sandboxed area for file I/O",
		"opt.server.run.static":      "Enforce static typing on program execution",
		"opt.server.run.superuser":   "Designate this user as a super-user with ROOT privileges",
		"opt.server.run.users":       "File with authentication JSON data",
		"opt.server.run.uuid":        "Sets the optional session UUID value",
		"opt.server.user.user":       "Username to create or update",
		"opt.server.user.pass":       "Password to assign to user",
		"opt.server.user.perms":      "Permissions to grant to user",

		"opt.table.create.file":        "File name containing JSON column info",
		"opt.table.delete.filter":      "Filter for rows to delete. If not specified, all rows are deleted",
		"opt.table.grant.permission":   "Permissions to set for this table updated",
		"opt.table.grant.user":         "User (if other than current user) to update",
		"opt.table.insert.file":        "File name containing JSON row info",
		"opt.table.list.no.row.counts": "If specified, listing does not include row counts",
		"opt.table.permission.user":    "User (if other than current user) to list)",
		"opt.table.permissions.user":   "If specified, list only this user",
		"opt.table.read.columns":       "List of optional column names to display; if not specified, all columns are returned",
		"opt.table.read.order.by":      "List of optional columns use to sort output",
		"opt.table.read.row.ids":       "Include the row UUID column in the output",
		"opt.table.read.row.numbers":   "Include the row number in the output",
		"opt.table.update.filter":      "Filter for rows to update. If not specified, all rows are updated",

		"parm.file":         "file",
		"parm.file.or.path": "file or path",
		"parm.address.port": "address:port",
		"parm.name":         "name",
		"parm.key":          "key",

		"parm.config.key.value": "key=value",

		"parm.sql.text": "sql-text",

		"parm.table.name":   "table-name",
		"parm.table.create": "table-name column:type [column:type...]",
		"parm.table.insert": "table-name [column=value...]",
		"parm.table.update": "table-name column=value [column=value...]",

		"error.array.bounds":           "array index out of bounds",
		"error.array.bracket":          "missing array bracket",
		"error.array.index":            "invalid array index",
		"error.array.value.type":       "wrong array value type",
		"error.arg.count":              "incorrect function argument count",
		"error.arg.type":               "incorrect function argument type",
		"error.argcheck.array":         "invalid ArgCheck array",
		"error.assert":                 "@assert error",
		"error.assert.testing":         "testing @assert failure",
		"error.asset":                  "no such asset",
		"error.assignment":             "missing '=' or ':='",
		"error.auth.type":              "invalid authentication type",
		"error.bit.shift":              "invalid bit shift specification",
		"error.block":                  "missing '{'",
		"error.block.end":              "missing '}'",
		"error.boolean.option":         "invalid boolean option value",
		"error.break.clause":           "invalid break clause",
		"error.bytecode.address":       "invalid bytecode address",
		"error.bytecode.not.found":     "unimplemented bytecode instruction",
		"error.cache.not.spec":         "cache size not specified",
		"error.call.frame":             "invalid call frame on stack",
		"error.cannot.delete.profile":  "cannot delete active profile",
		"error.case":                   "missing 'case'",
		"error.catch":                  "missing 'catch' clause",
		"error.channel.assignment":     "invalid use of assignment list for channel",
		"error.channel.not.open":       "channel not open",
		"error.cli.extra":              "unexpected text after command",
		"error.cli.option":             "unknown command line option",
		"error.cli.parms":              "too many parameters on command line",
		"error.cli.subcommand":         "unexpected parameters or invalid subcommand",
		"error.colon":                  "missing ':'",
		"error.column.count":           "incorrect number of columns",
		"error.column.name":            "invalid column name",
		"error.column.number":          "invalid column number",
		"error.column.width":           "invalid column width",
		"error.compiler":               "internal compiler error",
		"error.constant":               "invalid constant expression",
		"error.credentials":            "invalid credentials",
		"error.credentials.missing":    "no credentials provided",
		"error.db.closed":              "database client closed",
		"error.db.column.def":          "invalid database column definition",
		"error.db.result.type":         "invalid result set type",
		"error.db.rowset":              "invalid rowset value",
		"error.debug.service":          "cannot debug non-existent service",
		"error.debugger.cmd":           "invalid debugger command",
		"error.directive":              "invalid directive name",
		"error.directive.mode":         "directive invalid for mode",
		"error.div.zero":               "division by zero",
		"error.dup.column":             "duplicate column name",
		"error.dup.type":               "duplicate type name",
		"error.empty.column":           "empty column list",
		"error.entry.not.found":        "undefined entrypoint name",
		"error.equals":                 "missing '='",
		"error.expired":                "expired token",
		"error.expression":             "missing expression",
		"error.expression.term":        "missing term",
		"error.field.for.type":         "invalid field name for type",
		"error.field.not.found":        "unknown structure member",
		"error.file.id":                "invalid file identifier",
		"error.file.mode":              "invalid file open mode",
		"error.for.assignment":         "missing ':='",
		"error.for.body":               "for{} body empty",
		"error.for.exit":               "for{} has no exit",
		"error.for.init":               "missing for-loop initializer",
		"error.format.spec":            "invalid or unsupported format specification",
		"error.format.type":            "invalid output format type",
		"error.func.arg":               "invalid function argument",
		"error.func.call":              "invalid function invocation",
		"error.func.exists":            "function already defined",
		"error.func.name":              "invalid function name",
		"error.func.return.count":      "incorrect number of return values",
		"error.func.unused":            "function call used as parameter has unused error return value",
		"error.function":               "missing function",
		"error.function.body":          "missing function body",
		"error.function.name":          "missing function name",
		"error.function.list":          "missing function parameter list",
		"error.function.receiver":      "no function receiver",
		"error.function.return":        "missing function return type",
		"error.fucntion.values":        "missing return values",
		"error.general":                "general error",
		"error.go.error":               "Go routine {{name}} failed, {{err}}",
		"error.gremlin.client":         "invalid gremlin client",
		"error.http":                   "received HTTP",
		"error.identifier":             "invalid identifier",
		"error.identifier.not.found":   "unknown identifier",
		"error.import":                 "import not permitted inside a block or loop",
		"error.immutable.array":        "cannot change an immutable array",
		"error.immutable.map":          "cannot change an immutable map",
		"error.instruction":            "invalid instruction",
		"error.integer.option":         "invalid integer option value",
		"error.interface.imp":          "missing interface implementation",
		"error.invalid.blockquote":     "invalid block quote",
		"error.invalid.alignment.spec": "invalid alignment specification",
		"error.invalid.catch.set":      "invalid catch set {{index}}",
		"error.keyword.option":         "invalid option keyword",
		"error.list":                   "invalid list",
		"error.logger.confict":         "conflicting logger state",
		"error.logger.name":            "invalid logger name",
		"error.logon.endpoint":         "logon endpoint not found",
		"error.loop.control":           "loop control statement outside of for-loop",
		"error.loop.index":             "invalid loop index variable",
		"error.logon.server":           "no --logon-server specified",
		"error.map.key.type":           "wrong map key type",
		"error.map.value.type":         "wrong map value type",
		"error.media.type":             "invalid media type",
		"error.nil":                    "nil pointer reference",
		"error.not.assignment.list":    "not an assignment list",
		"error.not.channel":            "neither source or destination is a channel",
		"error.not.found":              "not found",
		"error.not.pointer":            "not a pointer",
		"error.not.service":            "not running as a service",
		"error.not.type":               "not a type",
		"error.opcode.defined":         "opcode already defined",
		"error.option.required":        "required option not found",
		"error.option.value":           "missing option value",
		"error.package.exists":         "cannot redefine existing package",
		"error.package.member":         "unknown package member",
		"error.package.name":           "invalid package name",
		"error.package.stmt":           "missing package statement",
		"error.panic":                  "Panic",
		"error.parens":                 "missing parenthesis",
		"error.parm.count":             "incorrect number of parameters",
		"error.parm.value.count":       "wrong number of parameter values",
		"error.pointer.type":           "invalid pointer type",
		"error.privilege":              "no privilege for operation",
		"error.profile.key":            "no such profile key",
		"error.profile.name":           "invalid configuration name",
		"error.profile.not.found":      "no such profile",
		"error.range":                  "invalid range",
		"error.readonly":               "invalid write to read-only item",
		"error.readonly.write":         "invalid write to read-only value",
		"error.reserved.name":          "reserved profile setting name",
		"error.rest.closed":            "rest client closed",
		"error.return.list":            "invalid return type list",
		"error.return.void":            "invalid return value for void function",
		"error.row.number":             "invalid row number",
		"error.sandbox.path":           "invalid sandbox path",
		"error.semicolon":              "missing ';'",
		"error.server.running":         "server already running as pid",
		"error.slice.index":            "invalid slice index",
		"error.spacing":                "invalid spacing value",
		"error.stack.underflow":        "stack underflow",
		"error.statement":              "missing statement",
		"error.statement.not.found":    "unrecognized statement",
		"error.step.type":              "invalid step type",
		"error.struct":                 "invalid result struct",
		"error.struct.type":            "unknown structure type",
		"error.symbol.exists":          "symbol already exists",
		"error.symbol.not.found":       "unknown symbol",
		"error.symbol.overflow":        "too many local symbols defined",
		"error.symbol.name":            "invalid symbol name",
		"error.table.closed":           "table closed",
		"error.table.processing":       "table processing",
		"error.template.name":          "invalid template name",
		"error.terminated":             "terminated with errors",
		"error.this":                   "invalid _this_ identifier",
		"error.timer":                  "invalid timer operation",
		"error.token.extra":            "unexpected token",
		"error.token.encryption":       "invalid token encryption",
		"error.type":                   "invalid or unsupported data type for this operation",
		"error.type.not.found":         "no such type",
		"error.try.stack":              "try/catch stack error",
		"error.type.check":             "invalid @type keyword",
		"error.type.def":               "missing type definition",
		"error.type.name":              "invalid type name",
		"error.type.spec":              "invalid type specification",
		"error.tx.actibe":              "transaction already active",
		"error.tx.not.active":          "no transaction active",
		"error.tx.not.found":           "no such transaction symbol",
		"error.url":                    "invalid URL path specification",
		"error.user.defined":           "user-supplied error",
		"error.user.not.found":         "no such user",
		"error.value":                  "invalid value",
		"error.value.extra":            "unexpected value",
		"error.var.args":               "invalid variable-argument operation",
		"error.var.type":               "invalid type for this variable",
		"error.version.parse":          "Unable to process version number {{v}; count={{c}}, err={{err}\n",
	},
}
