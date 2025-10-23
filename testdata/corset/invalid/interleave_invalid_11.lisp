;;error:4:22-23:invalid source column
(defcolumns (X :i16))
(defconst Y 1)
(definterleaved Z (X Y))
