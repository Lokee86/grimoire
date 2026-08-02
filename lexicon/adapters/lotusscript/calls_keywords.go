package main

var reservedCallWords = map[string]struct{}{
	"and": {}, "call": {}, "case": {}, "class": {}, "const": {}, "declare": {}, "dim": {},
	"do": {}, "else": {}, "elseif": {}, "end": {}, "error": {}, "exit": {},
	"for": {}, "forall": {}, "function": {}, "goto": {}, "if": {}, "is": {},
	"let": {}, "like": {}, "loop": {}, "mod": {}, "next": {}, "not": {}, "on": {},
	"global": {}, "open": {}, "option": {}, "or": {}, "preserve": {}, "print": {},
	"private": {}, "property": {}, "protected": {}, "public": {}, "redim": {},
	"resume": {}, "select": {}, "set": {}, "static": {}, "stop": {}, "sub": {}, "then": {},
	"type": {}, "use": {}, "uselsx": {}, "wend": {}, "while": {}, "with": {}, "xor": {},
}

var builtinFunctions = map[string]struct{}{
	"array": {}, "cbool": {}, "cbyte": {}, "ccur": {}, "cdate": {}, "cdat": {},
	"cdbl": {}, "cint": {}, "clng": {}, "createobject": {}, "csng": {}, "cstr": {},
	"chr": {}, "date": {}, "environ": {}, "evaluate": {}, "execute": {}, "format": {},
	"getobject": {}, "getthreadinfo": {}, "implode": {}, "instr": {}, "instrrev": {},
	"isarray": {}, "iselement": {}, "isempty": {}, "isnull": {}, "isobject": {},
	"join": {}, "lbound": {}, "lcase": {}, "left": {}, "len": {}, "lsi_info": {},
	"mid": {}, "msgbox": {}, "now": {}, "replace": {}, "right": {}, "run": {},
	"shell": {}, "split": {}, "strleft": {}, "strright": {}, "trim": {},
	"typename": {}, "ubound": {}, "ucase": {},
}
