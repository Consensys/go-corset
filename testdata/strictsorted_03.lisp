(defcolumns (P :binary@prove) (W0 :i16@prove) (W1 :i16@prove) (W2 :i4))
(defstrictsorted s1 P ((+ W2) (+ W1) (+ W0)))
