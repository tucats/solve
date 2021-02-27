package errors

import "errors"

// This contains the definitions for the Ego native errors, regardless
// of subsystem, etc.
// TODO introduce localized strings.

var ArgumentCountError = errors.New("incorrect function argument count")
var ArgumentTypeError = errors.New("incorrect function argument type")
var ArrayBoundsError = errors.New("array index out of bounds")
var AssertError = errors.New("@assert error")
var BlockQuoteError = errors.New("invalid block quote")
var CacheSizeNotSpecifiedError = errors.New("cache size not specified")
var CannotDeleteActiveProfile = errors.New("cannot delete active profile")
var ChannelNotOpenError = errors.New("channel not open")
var Continue = errors.New("continue")
var DatabaseClientClosedError = errors.New("database client closed")
var DivisionByZeroError = errors.New("division by zero")
var EmptyColumnListError = errors.New("empty column list")
var ExpiredTokenError = errors.New("expired token")
var FunctionAlreadyExistsError = errors.New("function already defined")
var GenericError = errors.New("general error")
var HTTPError = errors.New("received HTTP")
var ImmutableArrayError = errors.New("cannot change an immutable array")
var ImmutableMapError = errors.New("cannot change an immutable map")
var IncorrectColumnCountError = errors.New("incorred number of columns")
var IncorrectReturnValueCount = errors.New("incorrect number of return values")
var InternalCompilerError = errors.New("internal compiler error")
var InvalidAlignmentError = errors.New("invalid alignment specification")
var InvalidArgCheckError = errors.New("invalid ArgCheck array")
var InvalidArgTypeError = errors.New("function argument is of wrong type")
var InvalidArrayIndexError = errors.New("invalid array index")
var InvalidAuthenticationType = errors.New("invalid authentication type")
var InvalidBooleanValueError = errors.New("invalid boolean option value")
var InvalidBreakClauseError = errors.New("invalid break clause")
var InvalidBytecodeAddress = errors.New("invalid bytecode address")
var InvalidCallFrameError = errors.New("invalid call frame on stack")
var InvalidChannelError = errors.New("neither source or destination is a channel")
var InvalidChannelList = errors.New("invalid use of assignment list for channel")
var InvalidColumnNameError = errors.New("invalid column name")
var InvalidColumnNumberError = errors.New("invalid column number")
var InvalidColumnWidthError = errors.New("invalid column width")
var InvalidConstantError = errors.New("invalid constant expression")
var InvalidCredentialsError = errors.New("invalid credentials")
var InvalidDebugCommandError = errors.New("invalid debugger command")
var InvalidDirectiveError = errors.New("invalid directive name")
var InvalidFieldError = errors.New("invalid field name for type")
var InvalidFunctionArgument = errors.New("invalid function argument")
var InvalidFunctionCall = errors.New("invalid function invocation")
var InvalidFunctionCallError = errors.New("invalid function call")
var InvalidFunctionName = errors.New("invalid function name")
var InvalidGremlinClientError = errors.New("invalid gremlin client")
var InvalidIdentifierError = errors.New("invalid identifier")
var InvalidImportError = errors.New("import not permitted inside a block or loop")
var InvalidInstructionError = errors.New("invalid instruction")
var InvalidIntegerError = errors.New("invalid integer option value")
var InvalidKeywordError = errors.New("invalid option keyword")
var InvalidListError = errors.New("invalid list")
var InvalidLoggerName = errors.New("invalid logger name")
var InvalidLoopControlError = errors.New("loop control statement outside of for-loop")
var InvalidLoopIndexError = errors.New("invalid loop index variable")
var InvalidOutputFormatErr = errors.New("invalid output format specified")
var InvalidOutputFormatError = errors.New("invalid output format")
var InvalidPackageName = errors.New("invalid package name")
var InvalidPointerTypeError = errors.New("invalid pointer type")
var InvalidRangeError = errors.New("invalid range")
var InvalidResultSetTypeError = errors.New("invalid result set type")
var InvalidReturnTypeList = errors.New("invalid return type list")
var InvalidReturnValueError = errors.New("invalid return value for void function")
var InvalidRowNumberError = errors.New("invalid row number")
var InvalidRowSetError = errors.New("invalid rowset value")
var InvalidSliceIndexError = errors.New("invalid slice index")
var InvalidSpacingError = errors.New("invalid spacing value")
var InvalidStepType = errors.New("invalid step type")
var InvalidStructError = errors.New("invalid result struct")
var InvalidSymbolError = errors.New("invalid symbol name")
var InvalidTemplateName = errors.New("invalid template name")
var InvalidThisError = errors.New("invalid _this_ identifier")
var InvalidTimerError = errors.New("invalid timer operation")
var InvalidTokenEncryption = errors.New("invalid token encryption")
var InvalidTypeCheckError = errors.New("invalid @type keyword")
var InvalidTypeError = errors.New("invalid or unsupported data type for this operation")
var InvalidTypeNameError = errors.New("invalid type name")
var InvalidTypeSpecError = errors.New("invalid type specification")
var InvalidValueError = errors.New("invalid value")
var InvalidVarTypeError = errors.New("invalid type for this variable")
var InvalidFormatVerbError = errors.New("invalid or unsupported format specification")
var InvalidfileIdentifierError = errors.New("invalid file identifier")
var LogonEndpointError = errors.New("logon endpoint not found")
var LoopBodyError = errors.New("for{} body empty")
var LoopExitError = errors.New("for{} has no exit")
var MissingAssignmentError = errors.New("missing '=' or ':='")
var MissingBlockError = errors.New("missing '{'")
var MissingBracketError = errors.New("missing array bracket")
var MissingCaseError = errors.New("missing 'case'")
var MissingCatchError = errors.New("missing 'catch' clause")
var MissingColonError = errors.New("missing ':'")
var MissingEndOfBlockError = errors.New("missing '}'")
var MissingEqualError = errors.New("missing '='")
var MissingForLoopInitializerError = errors.New("missing for-loop initializer")
var MissingFunctionBodyError = errors.New("missing function body")
var MissingFunctionTypeError = errors.New("missing function return type")
var MissingLoopAssignmentError = errors.New("missing ':='")
var MissingOptionValueError = errors.New("missing option value")
var MissingOutputTypeError = errors.New("missing output format type")
var MissingPackageStatement = errors.New("missing package statement")
var MissingParameterList = errors.New("missing function parameter list")
var MissingParenthesisError = errors.New("missing parenthesis")
var MissingReturnValues = errors.New("missing return values")
var MissingSemicolonError = errors.New("missing ';'")
var MissingTermError = errors.New("missing term")
var NilPointerReferenceError = errors.New("nil pointer reference")
var NoCredentialsError = errors.New("no credentials provided")
var NoFunctionReceiver = errors.New("no function receiver")
var NoLogonServerError = errors.New("no --logon-server specified")
var NoPrivilegeForOperationError = errors.New("no privilege for operation")
var NoSuchAsset = errors.New("no such asset")
var NoSuchDebugService = errors.New("cannot debug non-existent service")
var NoSuchProfile = errors.New("no such profile")
var NoSuchUserError = errors.New("no such user")
var NotAPointer = errors.New("not a pointer")
var NoTransactionActiveError = errors.New("no transaction active")
var NotAServiceError = errors.New("not running as a service")
var NotATypeError = errors.New("not a type")
var NotAnLValueListError = errors.New("not an lvalue list")
var OpcodeAlreadyDefinedError = errors.New("opcode already defined")
var PackageRedefinitionError = errors.New("cannot redefine existing package")
var Panic = errors.New("Panic")
var ReadOnlyError = errors.New("invalid write to read-only item")
var ReadOnlyValueError = errors.New("invalid write to read-only value")
var RegisterAddressError = errors.New("internal register address error")
var RequiredNotFoundError = errors.New("required option not found")
var ReservedProfileSetting = errors.New("reserved profile setting name")
var RestClientClosedError = errors.New("rest client closed")
var ServerAlreadyRunning = errors.New("server already running as pid")
var SignalDebugger = errors.New("signal")
var StackUnderflowError = errors.New("stack underflow")
var StepOver = errors.New("step-over")
var Stop = errors.New("stop")
var SymbolExistsError = errors.New("symbol already exists")
var TableClosedError = errors.New("table closed")
var TableErrorPrefix = errors.New("table processing")
var TerminatedWithErrors = errors.New("terminated with errors")
var TestingAssertError = errors.New("testing @assert failure")
var TooManyParametersError = errors.New("too many parameters on command line")
var TooManyLocalSymbols = errors.New("too many local symbols defined")
var TooManyReturnValues = errors.New("too many return values")
var TransactionAlreadyActive = errors.New("transaction already active")
var TryCatchMismatchError = errors.New("try/catch stack error")
var UnexpectedParametersError = errors.New("unexpected parameters or invalid subcommand")
var UnepectedTextAfterCommandError = errors.New("unexpected text after command")
var UnexpectedTokenError = errors.New("unexpected token")
var UnexpectedValueError = errors.New("unexpected value")
var UnimplementedInstructionError = errors.New("unimplemented bytecode instruction")
var UnknownIdentifierError = errors.New("unknown identifier")
var UnknownMemberError = errors.New("unknown structure member")
var UnknownOptionError = errors.New("unknown command line option")
var UnknownPackageMemberError = errors.New("unknown package member")
var UnknownSymbolError = errors.New("unknown symbol")
var UnknownTypeError = errors.New("unknown structure type")
var UnrecognizedStatementError = errors.New("unrecognized statement")
var UserError = errors.New("user-supplied error")
var VarArgError = errors.New("invalid variable-argument operation")
var WrongArrayValueType = errors.New("wrong array value type")
var WrongMapKeyType = errors.New("wrong map key type")
var WrongMapValueType = errors.New("wrong map value type")
var WrongModeError = errors.New("directive invalid for mode")
var WrongParameterCountError = errors.New("incorrect number of parameters")
