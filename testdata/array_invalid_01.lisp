;;error:4:24-31:expected array column
(defcolumns (BIT :i16@loob))

(defconstraint bits () [BIT 1])
