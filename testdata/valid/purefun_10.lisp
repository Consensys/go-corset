(defcolumns (X :i16))
(defpurefun (idf x) x)
(defcomputedcolumn (Y :i16) (idf X))
