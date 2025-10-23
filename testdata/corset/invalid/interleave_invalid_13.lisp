;;error:3:20-21:recursive definition
(defcolumns (X :i16))
(definterleaved Z (Z X))
