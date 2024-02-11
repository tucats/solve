package errors

// This contains the definitions for the Ego native errors, regardless
// of subsystem, etc.
// TODO introduce localized strings.

// Return values used to signal flow change.
// THESE SHOULD NOT BE LOCALIZED.

var ErrContinue = Message("_continue")
var ErrSignalDebugger = Message("_signal")
var ErrStepOver = Message("_step-over")
var ErrStop = Message("_stop")
var ErrExit = Message("_exit")

// Return values reflecting runtime error conditions.

var ErrAlignment = Message("invalid.alignment.spec")
var ErrArgumentCount = Message("arg.count")
var ErrArgumentType = Message("arg.type")
var ErrArgumentTypeCheck = Message("argcheck.array")
var ErrArrayBounds = Message("array.bounds")
var ErrArrayIndex = Message("array.index")
var ErrAssert = Message("assert")
var ErrBlockQuote = Message("invalid.blockquote")
var ErrCacheSizeNotSpecified = Message("cache.not.spec")
var ErrCannotDeleteActiveProfile = Message("cannot.delete.profile")
var ErrCertificateParseError = Message("cert.parse")
var ErrChannelNotOpen = Message("channel.not.open")
var ErrChildTimeout = Message("child.timeout")
var ErrColumnCount = Message("column.count")
var ErrConditionalBool = Message("conditional.bool")
var ErrDatabaseClientClosed = Message("db.closed")
var ErrDeferOutsideFunction = Message("defer.outside")
var ErrDivisionByZero = Message("div.zero")
var ErrDuplicateColumnName = Message("dup.column")
var ErrDuplicateTypeName = Message("dup.type")
var ErrEmptyColumnList = Message("empty.column")
var ErrExpiredToken = Message("expired")
var ErrExtension = Message("extension")
var ErrFunctionAlreadyExists = Message("func.exists")
var ErrFunctionReturnedVoid = Message("func.void")
var ErrGeneric = Message("general")
var ErrHTTP = Message("http")
var ErrImmutableArray = Message("immutable.array")
var ErrImmutableMap = Message("immutable.map")
var ErrImportNotCached = Message("import.not.found")
var ErrInternalCompiler = Message("compiler")
var ErrInvalidArgumnetList = Message("arg.list")
var ErrInvalidAuthenticationType = Message("auth.type")
var ErrInvalidAuto = Message("invalid.auto")
var ErrInvalidBitShift = Message("bit.shift")
var ErrInvalidBitSize = Message("bit.size")
var ErrInvalidBooleanValue = Message("boolean.option")
var ErrInvalidBreakClause = Message("break.clause")
var ErrInvalidBytecodeAddress = Message("bytecode.address")
var ErrInvalidCallFrame = Message("call.frame")
var ErrInvalidChannel = Message("not.channel")
var ErrInvalidChannelList = Message("channel.assignment")
var ErrInvalidColumnDefinition = Message("db.column.def")
var ErrInvalidColumnName = Message("column.name")
var ErrInvalidColumnNumber = Message("column.number")
var ErrInvalidColumnWidth = Message("column.width")
var ErrInvalidConfigName = Message("profile.name")
var ErrInvalidConstant = Message("constant")
var ErrInvalidCredentials = Message("credentials")
var ErrInvalidDebugCommand = Message("debugger.cmd")
var ErrInvalidDirective = Message("directive")
var ErrInvalidEndPointString = Message("endpoint")
var ErrInvalidField = Message("field.for.type")
var ErrInvalidFileMode = Message("file.mode")
var ErrInvalidFormatVerb = Message("format.spec")
var ErrInvalidFunctionArgument = Message("func.arg")
var ErrInvalidFunctionCall = Message("func.call")
var ErrInvalidFunctionTypeCall = Message("func.type.call")
var ErrInvalidFunctionName = Message("func.name")
var ErrInvalidIdentifier = Message("identifier")
var ErrInvalidImport = Message("import")
var ErrInvalidInstruction = Message("instruction")
var ErrInvalidInteger = Message("integer.option")
var ErrInvalidKeyword = Message("keyword.option")
var ErrInvalidList = Message("list")
var ErrInvalidLoggerName = Message("logger.name")
var ErrInvalidLoopControl = Message("loop.control")
var ErrInvalidLoopIndex = Message("loop.index")
var ErrInvalidMediaType = Message("media.type")
var ErrInvalidOperand = Message("operand")
var ErrInvalidOutputFormat = Message("format.type")
var ErrInvalidPackageName = Message("package.name")
var ErrInvalidPermission = Message("permission.name")
var ErrInvalidPointerType = Message("pointer.type")
var ErrInvalidRange = Message("range")
var ErrInvalidResultSetType = Message("db.result.type")
var ErrInvalidReturnTypeList = Message("return.list")
var ErrInvalidNamedReturnValues = Message("named.return.values")
var ErrInvalidReturnValue = Message("return.void")
var ErrInvalidReturnValues = Message("named.return.values")
var ErrInvalidRowNumber = Message("row.number")
var ErrInvalidRowSet = Message("db.rowset")
var ErrInvalidSandboxPath = Message("sandbox.path")
var ErrInvalidScopeLevel = Message("scope.invalid")
var ErrInvalidSliceIndex = Message("slice.index")
var ErrInvalidSpacing = Message("spacing")
var ErrInvalidStepType = Message("step.type")
var ErrInvalidStruct = Message("struct")
var ErrInvalidStructOrPackage = Message("invalid.struct.or.package")
var ErrInvalidSymbolName = Message("symbol.name")
var ErrInvalidTemplateName = Message("template.name")
var ErrInvalidThis = Message("this")
var ErrInvalidTimer = Message("timer")
var ErrInvalidTokenEncryption = Message("token.encryption")
var ErrInvalidType = Message("type")
var ErrInvalidTypeCheck = Message("type.check")
var ErrInvalidTypeName = Message("type.name")
var ErrInvalidTypeSpec = Message("type.spec")
var ErrInvalidUnwrap = Message("invalid.unwrap")
var ErrInvalidURL = Message("url")
var ErrInvalidValue = Message("value")
var ErrInvalidVarType = Message("var.type")
var ErrInvalidVariableArguments = Message("var.args")
var ErrInvalidfileIdentifier = Message("file.id")
var ErrLoggerConflict = Message("logger.conflict")
var ErrLogonEndpoint = Message("logon.endpoint")
var ErrLoopBody = Message("for.body")
var ErrLoopExit = Message("for.exit")
var ErrMissingAssignment = Message("assignment")
var ErrMissingBlock = Message("block")
var ErrMissingBracket = Message("array.bracket")
var ErrMissingCase = Message("case")
var ErrMissingCatch = Message("catch")
var ErrMissingColon = Message("colon")
var ErrMissingEndOfBlock = Message("block.end")
var ErrMissingEqual = Message("equals")
var ErrMissingExpression = Message("expression")
var ErrMissingForLoopInitializer = Message("for.init")
var ErrMissingFunction = Message("function")
var ErrMissingFunctionBody = Message("function.body")
var ErrMissingFunctionName = Message("function.name")
var ErrMissingFunctionType = Message("function.return")
var ErrMissingInterface = Message("interface.imp")
var ErrMissingLoggerName = Message("logger.name")
var ErrMissingLoopAssignment = Message("for.assignment")
var ErrMissingOptionValue = Message("option.value")
var ErrMissingOutputType = Message("format.type")
var ErrMissingPackageName = Message("package.name")
var ErrMissingPackageStatement = Message("package.stmt")
var ErrMissingParameterList = Message("function.list")
var ErrMissingParenthesis = Message("parens")
var ErrMissingPrintItems = Message("print.items")
var ErrMissingReturnValues = Message("function.values")
var ErrMissingSemicolon = Message("semicolon")
var ErrMissingStatement = Message("statement")
var ErrMissingSymbol = Message("symbol.name")
var ErrMissingTerm = Message("expression.term")
var ErrMissingType = Message("type.def")
var ErrNilPointerReference = Message("nil")
var ErrNoCredentials = Message("credentials.missing")
var ErrNoDatabase = Message("no.database")
var ErrNoInfo = Message("no.info")
var ErrNoFunctionReceiver = Message("function.receiver")
var ErrNoLogonServer = Message("logon.server")
var ErrNoMainPackage = Message("no.main.package")
var ErrNoPrivilegeForOperation = Message("privilege")
var ErrNoSuchAsset = Message("asset")
var ErrNoSuchDebugService = Message("debug.service")
var ErrNoSuchDSN = Message("dsn.not.found")
var ErrNoSuchProfile = Message("profile.not.found")
var ErrNoSuchProfileKey = Message("profile.key")
var ErrNoSuchTXSymbol = Message("tx.not.found")
var ErrNoSuchUser = Message("user.not.found")
var ErrNoSymbolTable = Message("no.symbol.table")
var ErrNoTransactionActive = Message("tx.not.active")
var ErrNotAPointer = Message("not.pointer")
var ErrNotAService = Message("not.service")
var ErrNotAType = Message("not.type")
var ErrNotAnLValueList = Message("not.assignment.list")
var ErrNotFound = Message("not.found")
var ErrOpcodeAlreadyDefined = Message("opcode.defined")
var ErrPackageRedefinition = Message("package.exists")
var ErrPanic = Message("panic")
var ErrReadOnly = Message("readonly")
var ErrReadOnlyAddressable = Message("readonly.addressable")
var ErrReadOnlyValue = Message("readonly.write")
var ErrRequiredNotFound = Message("option.required")
var ErrReservedProfileSetting = Message("reserved.name")
var ErrRestClientClosed = Message("rest.closed")
var ErrReturnValueCount = Message("func.return.count")
var ErrServerAlreadyRunning = Message("server.running")
var ErrStackUnderflow = Message("stack.underflow")
var ErrSymbolNotExported = Message("symbol.not.exported")
var ErrSymbolExists = Message("symbol.exists")
var ErrTableClosed = Message("table.closed")
var ErrTableErrorPrefix = Message("table.processing")
var ErrTerminatedWithErrors = Message("terminated")
var ErrTestingAssert = Message("assert.testing")
var ErrTooManyLocalSymbols = Message("symbol.overflow")
var ErrTooManyParameters = Message("cli.parms")
var ErrTooManyReturnValues = Message("func.return.count")
var ErrTransactionAlreadyActive = Message("tx.active")
var ErrTryCatchMismatch = Message("try.stack")
var ErrTypeMismatch = Message("type.mismatch")
var ErrUndefinedEntrypoint = Message("entry.not.found")
var ErrUnexpectedParameters = Message("cli.subcommand")
var ErrUnexpectedTextAfterCommand = Message("cli.extra")
var ErrUnexpectedToken = Message("token.extra")
var ErrUnexpectedValue = Message("value.extra")
var ErrUnimplementedInstruction = Message("bytecode.not.found")
var ErrUnknownIdentifier = Message("identifier.not.found")
var ErrUnknownMember = Message("field.not.found")
var ErrUnknownOption = Message("cli.option")
var ErrUnknownPackageMember = Message("package.member")
var ErrUnknownSymbol = Message("symbol.not.found")
var ErrUnknownType = Message("type.not.found")
var ErrUnrecognizedCommand = Message("cli.command.not.found")
var ErrUnrecognizedStatement = Message("statement.not.found")
var ErrUnsupportedOnOS = Message("unsupported.on.os")
var ErrUnusedErrorReturn = Message("func.unused")
var ErrUserDefined = Message("user.defined")
var ErrWrongArrayValueType = Message("array.value.type")
var ErrWrongMapKeyType = Message("map.key.type")
var ErrWrongMapValueType = Message("map.value.type")
var ErrWrongMode = Message("directive.mode")
var ErrWrongParameterCount = Message("parm.count")
var ErrWrongParameterValueCount = Message("parm.value.count")
var ErrWrongUserUpdatedCount = Message("user.count")
