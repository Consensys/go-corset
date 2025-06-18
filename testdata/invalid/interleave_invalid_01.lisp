;;error:3:17-18:symbol Z already declared
(defcolumns (X :i16) (Y :i16) (Z :i16))
(definterleaved Z (X Y))
