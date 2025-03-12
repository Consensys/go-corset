;;error:4:24-27:expected loobean constraint (found (u16@loob)[1:4])
(defcolumns (BIT :i16@loob :array [4]))

(defconstraint bits () BIT)
