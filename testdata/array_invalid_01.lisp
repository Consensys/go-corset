;;error:4:24-30:not an array column
(defcolumns (BIT :@loob))

(defconstraint bits () [BIT 1])
