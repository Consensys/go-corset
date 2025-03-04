(defpurefun (prev X) (shift X -1))

(defcolumns (P :binary@prove) (W0 :i16@prove))
(defstrictsorted s1 (prev P) ((+ W0)))
