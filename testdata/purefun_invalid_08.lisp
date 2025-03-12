;;error:6:22-28:expected loobean constraint (found u16)
(defcolumns (X :i16@loob) (Y :i16))
(defpurefun (fd x) x)

(defconstraint c1 () (fd X))
(defconstraint c2 () (fd Y))
