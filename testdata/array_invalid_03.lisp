;;error:4:29-30:expected constant array index
(defcolumns X (BIT :i16@loob :array [4]))

(defconstraint bits () [BIT X])
