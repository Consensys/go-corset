;;
(module m1)

(defcolumns (X :i16) (Y :i16))
(defconstraint test () (== X Y))

(module m1)
(defconstraint test () (!= X Y))
