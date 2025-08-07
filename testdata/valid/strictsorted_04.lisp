(defcolumns (P :i3) (Q :i3) (W0 :i16@prove))
(defstrictsorted s1 (+ P Q) ((↓ W0)))
(defstrictsorted s2 (+ P Q) ((+ W0)))
