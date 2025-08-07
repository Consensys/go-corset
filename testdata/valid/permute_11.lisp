;; this test is all about expanded traces
(defcolumns (W0 :byte) (W1 :byte))
(defpermutation (A B) ((+ W1) W0))
