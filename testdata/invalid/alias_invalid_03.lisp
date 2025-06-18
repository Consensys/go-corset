;;error:3:11-12:symbol already exists
(defcolumns (X :i16) (Y :i16))
(defalias X Y)
