;;error:7:26-27:expected bool, found u1
;;error:7:28-29:expected bool, found u4
(defcolumns (X :i1) (Y :i1) (A :i4) (B :i4))
(definterleaved Z (X Y))
(definterleaved C (A B))
;;
(defconstraint c1 () (if Z C))
