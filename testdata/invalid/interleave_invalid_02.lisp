;;error:3:17-19:malformed target column
(defcolumns (X :i16) (Y :i16))
(definterleaved () (X Y))
