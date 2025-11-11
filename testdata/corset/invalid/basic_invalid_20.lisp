;;error:8:16-20:duplicate handle
(module m1)

(defcolumns (X :i16) (Y :i16))
(defconstraint test () (== X Y))

(module m1)
(defconstraint test () (!= X Y))
