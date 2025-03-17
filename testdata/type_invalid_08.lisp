;;error:6:22-30:expected loobean constraint (found u4)
(defcolumns (X :i1) (Y :i1) (A :i4) (B :i4))
(definterleaved Z (X Y))
(definterleaved C (A B))
;;
(defconstraint c1 () (if Z C))
