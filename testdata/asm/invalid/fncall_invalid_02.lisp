;;error:3:19-20 too many arguments for function "f"
(defcolumns (X :i16) (Y :i16))
(defcall (Y) f (X X))
