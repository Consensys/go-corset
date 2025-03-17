;;error:4:24-27:expected loobean constraint (found (u16)[1:4])
(defcolumns (BIT :i16 :array [4]))

(defconstraint bits () BIT)
