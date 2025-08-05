;;error:4:20-28:symbol Y already declared
(defcolumns (X :i16))
(defcolumns (Y :i24))
(defcomputedcolumn (Y :i24) (+ X 1))
