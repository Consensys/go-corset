;;error:2:1-2:blah
(defcolumns (X :@loob) Y)
(defpurefun (id x) x)

(defconstraint c1 () (id X))
(defconstraint c2 () (id Y))
