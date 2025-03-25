;;error:4:24-27:expected bool, found (u16)[1:4]
(defcolumns (BIT :i16 :array [4]))

(defconstraint bits () BIT)
