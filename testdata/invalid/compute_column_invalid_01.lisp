;;error:3:20-21:computed column is not of the right format
(defcolumns (X :i16))
(defcomputedcolumn Y (+ X 1))