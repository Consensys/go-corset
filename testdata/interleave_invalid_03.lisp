;;error:3:20-23:malformed source column
(defcolumns (X :i16) (Y :i16))
(definterleaved Z ((X) Y))
