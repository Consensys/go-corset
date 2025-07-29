(defcolumns (X :i8 :array [0:1]) (Y :i16))
;;
(defconstraint c1 () (== Y (:: [X 1] [X 0])))
