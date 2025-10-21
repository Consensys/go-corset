;;error:3:17-19:empty column declaration
(defcolumns (X :i16) (Y :i16))
(definterleaved () (X Y))
