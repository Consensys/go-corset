;;error:5:22-28:expected bool, found u16
(defcolumns (X :i16))
(defpurefun (fd x) x)

(defconstraint c1 () (fd X))
