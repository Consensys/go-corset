;;error:4:11-13:symbol already exists
(defcolumns (X :i16) (Y :i16))
(defalias CT X)
(defalias CT Y)
