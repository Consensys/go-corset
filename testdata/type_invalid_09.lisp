;;error:6:26-27:invalid condition (neither loobean nor boolean)
(defcolumns (X :i1) (Y :i1) (A :i4@loob) (B :i4@loob))
(definterleaved Z (X Y))
(definterleaved C (A B))
;;
(defconstraint c1 () (if Z C))
