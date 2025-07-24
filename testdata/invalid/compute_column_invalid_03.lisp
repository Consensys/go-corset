;;error:3:14-15:malformed target columns
(defcolumns (X :i16))
(defcolumns (Y :i24))
(defcomputedcolumn (Y) (+ X 1))